package replication

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"math"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/spf13/afero"
)

const (
	constWriteErrorGeneral = 1

	headerSize = 16 + 4 + 4 + 8 // position 16 bytes + size 4 bytes + crc 4 bytes + zero 8 bytes
)

var (
	fileNameLength   = len("wal_0000000000000001.log")
	startingPosition = Position{1, 0}
)

type RoughReader interface {
}

type WalOptions struct {
	FS afero.Fs
}

type Wal struct {
	opt WalOptions

	currentPosition Position

	rwlock     sync.RWMutex
	writeQueue chan struct {
		Data  []byte
		ResCh chan struct {
			Code int
			Pos  Position
		}
	}

	openFile io.WriteCloser
}

func NewWal(opt WalOptions) *Wal {
	w := &Wal{
		opt: opt,

		currentPosition: startingPosition, // initial position if file not found

		writeQueue: make(chan struct {
			Data  []byte
			ResCh chan struct {
				Code int
				Pos  Position
			}
		}, 128),
	}
	return w
}

func (w *Wal) Startup() error {
	// read existing file
	if err := w.readAndCheckExistingFile(); err != nil {
		return err
	}

	if err := w.initializeFile(); err != nil {
		return err
	}

	go w.WriteLoop()

	return nil
}

func (w *Wal) Shutdown() error {
	//TODO close writeQueue

	if w.openFile != nil {
		_ = w.openFile.Close()
	}

	return nil
}

// RoughReader returns a RoughReader for wal replaying
// If wal logs are missing between pos and earliest wal file, nil is returned as reader w/o any error
// If preparing reader failed, an error is returned
func (w *Wal) RoughReader(pos Position) (RoughReader, error) {
	//TODO create rough reader
	// before reading file, record ending position
	// then read from given position to the ending position
	// any error is accepted beyond the range
	return nil, nil
}

func (w *Wal) Write(data []byte) (Position, error) {
	req := struct {
		Data  []byte
		ResCh chan struct {
			Code int
			Pos  Position
		}
	}{
		Data: w.prepareWriteEntry(data),
		ResCh: make(chan struct {
			Code int
			Pos  Position
		}, 1),
	}
	w.writeQueue <- req
	res := <-req.ResCh
	if res.Code != 0 {
		return Position{}, fmt.Errorf("write error: %d", res.Code)
	} else {
		return res.Pos, nil
	}
}

func (w *Wal) CurrentPosition() (Position, error) {
	w.rwlock.RLock()
	defer w.rwlock.RUnlock()
	return w.currentPosition, nil
}

func (w *Wal) WriteLoop() {
	exist := false
	for !exist {
		req, ok := <-w.writeQueue
		if !ok {
			exist = true
			continue
		}
		f := func() {
			w.rwlock.Lock()
			defer w.rwlock.Unlock()

			pos, err := w.doWrite(req.Data)
			if err != nil {
				//FIXME print error log
				req.ResCh <- struct {
					Code int
					Pos  Position
				}{Code: constWriteErrorGeneral, Pos: Position{}}
			} else {
				req.ResCh <- struct {
					Code int
					Pos  Position
				}{Code: 0, Pos: pos}
			}
		}
		f()
	}
}

func (w *Wal) doWrite(data []byte) (Position, error) {
	// next position
	newPos := w.nextPos()
	// prepare position
	copy(data[0:16], newPos.ToBytes())
	// write file
	if _, err := w.openFile.Write(data); err != nil {
		return Position{}, err
	}
	// commit current position
	w.commitPos(newPos)
	return newPos, nil
}

func (w *Wal) prepareWriteEntry(data []byte) []byte {
	// prepare complete write entry except position field
	newBuf := make([]byte, len(data)+headerSize)
	// size
	binary.BigEndian.PutUint32(newBuf[16:16+4], uint32(len(data)))
	// crc
	binary.BigEndian.PutUint32(newBuf[16+4:16+4+4], adler32.Checksum(data))
	// data
	copy(newBuf[headerSize:], data)
	return newBuf
}

func (w *Wal) nextPos() Position {
	oldPos := w.currentPosition
	newPos := Position{
		oldPos[0], oldPos[1] + 1,
	}
	return newPos
}

func (w *Wal) commitPos(pos Position) {
	w.currentPosition = pos
}

func (w *Wal) initializeFile() error {
	oldPos := w.currentPosition
	newPos := Position{
		oldPos[0], oldPos[1],
	}
	f, err := w.opt.FS.OpenFile(fmt.Sprintf("wal_%016x.log", newPos[0]), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.openFile = f
	w.currentPosition = newPos
	return nil
}

func (w *Wal) readAndCheckExistingFile() error {
	files, err := afero.ReadDir(w.opt.FS, ".")
	if err != nil {
		return err
	}
	m := map[string]struct{}{}
	for _, fi := range files {
		if fi.IsDir() {
			continue
		}
		if len(fi.Name()) == fileNameLength {
			m[fi.Name()] = struct{}{}
		}
	}

	// no file found, fresh folder
	if len(m) == 0 {
		return nil
	}

	// prepare file list
	var fileSlice []struct {
		Name  string
		Index int64
	}
	for k, _ := range m {
		var idx int64
		if n, err := fmt.Sscanf(k, "wal_%016x.log", &idx); err != nil {
			//FIXME log unrecognized file name
			continue
		} else if n != 1 {
			//FIXME log no index captured from the file name
			continue
		}
		fileSlice = append(fileSlice, struct {
			Name  string
			Index int64
		}{
			Name:  k,
			Index: idx,
		})
	}
	slices.SortFunc(fileSlice, func(a, b struct {
		Name  string
		Index int64
	}) int {
		return strings.Compare(a.Name, b.Name)
	})
	// check ascending
	for idx := 1; idx < len(fileSlice); idx++ {
		if fileSlice[idx].Index-fileSlice[idx-1].Index == 1 {
			continue
		} else {
			return errors.New("wal file names are not continuous")
		}
	}
	// read last file and prepare current position
	last := fileSlice[len(fileSlice)-1]
	f, err := w.opt.FS.Open(last.Name)
	if err != nil {
		return err
	}
	defer func(f afero.File) {
		_ = f.Close()
	}(f)
	wfr := &walFileReader{r: f, idx: last.Index}
	pos, err := wfr.readLastPos()
	if err != nil {
		return err
	}
	w.currentPosition = pos

	return nil
}

type walFileReader struct {
	r   io.Reader
	idx int64
}

func (w *walFileReader) readLastPos() (Position, error) {
	header := make([]byte, headerSize)
	var pos = Position{w.idx, 0}
	for {
		_, err := io.ReadFull(w.r, header)
		if err == io.EOF {
			//FIXME need support corruption file
			return Position{w.idx, pos[1]}, nil
		}
		size := binary.BigEndian.Uint32(header[16:])
		buf := make([]byte, size)
		_, err = io.ReadFull(w.r, buf)
		if err != nil {
			//FIXME need support corruption file
			return Position{}, err
		}
		//FIXME need check crc32
		pos = Position{w.idx, int64(binary.BigEndian.Uint64(header[8:]) & math.MaxInt64)}
	}
}
