package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"slices"
)

const (
	MaxConfigurationGenerated = 3_000_000 // target 3_000_000
	MaxClientGenerated        = 1_000_000 // target 1_000_000
)

type ConfigurationList struct {
	List []struct {
		Group string
		Key   string
	}
	Clients []struct {
		Consuming []struct {
			Group string
			Key   string
		}
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func main() {
	var rnd = rand.New(rand.NewPCG(2024, 1003))

	// type1: 3m configure keys
	cfgList := DataGenerateConfigureKeys(rnd)
	// type2: 1m clients and random 100 configures consuming
	clients := DataGenerateClients(rnd, cfgList)
	log.Println("generation completed.")

	// generate statistics
	GenerateStatistics(cfgList, clients)
	PrintStatistics(cfgList, clients)

	// write files
	writeFiles(cfgList, clients)
}

func writeFiles(list []*struct {
	Group string
	Key   string
	Count int
}, clients []*struct {
	Consume []*struct {
		Group string
		Key   string
		Idx   int
	}
}) {
	r := &ConfigurationList{}
	for _, v := range list {
		r.List = append(r.List, struct {
			Group string
			Key   string
		}{Group: v.Group, Key: v.Key})
	}
	for _, v := range clients {
		item := struct {
			Consuming []struct {
				Group string
				Key   string
			}
		}{}
		for _, vv := range v.Consume {
			item.Consuming = append(item.Consuming, struct {
				Group string
				Key   string
			}{Group: vv.Group, Key: vv.Key})
		}
		r.Clients = append(r.Clients, item)
	}

	f, err := os.OpenFile("dump.data", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("close file failed:", err)
		}
	}(f)
	log.Println("start dumping data to files...")
	if err := json.NewEncoder(f).Encode(r); err != nil {
		log.Fatal(err)
	}
	log.Println("dump completed.")
}

func PrintStatistics(list []*struct {
	Group string
	Key   string
	Count int
}, clients []*struct {
	Consume []*struct {
		Group string
		Key   string
		Idx   int
	}
}) {
	var totalConsuming int
	for _, item := range list {
		totalConsuming += item.Count
	}
	log.Println("total consuming:", totalConsuming)

	log.Println("first configure listener count:", list[0].Count)
	log.Println("first 1100 items ratio list:")
	for i := 0; i < 1100; i++ {
		log.Printf("kth:%d ratio:%f%%\n", i+1, float64(list[i].Count)*100/MaxClientGenerated)
	}

	var hasZeroCount bool
	for idx, item := range list {
		if item.Count == 0 {
			log.Println("count check alert: has zero count at index:", idx)
			hasZeroCount = true
		}
	}
	if !hasZeroCount {
		log.Println("count check pass: no zero count configure found")
	}
	log.Printf("last item ratio:%f%%\n", float64(list[len(list)-1].Count)*100/MaxClientGenerated)
}

func GenerateStatistics(list []*struct {
	Group string
	Key   string
	Count int
}, clients []*struct {
	Consume []*struct {
		Group string
		Key   string
		Idx   int
	}
}) {
	for _, client := range clients {
		for _, c := range client.Consume {
			list[c.Idx].Count = list[c.Idx].Count + 1
		}
	}
	// sort by count
	slices.SortFunc(list, func(a, b *struct {
		Group string
		Key   string
		Count int
	}) int {
		return b.Count - a.Count
	})
}

func DataGenerateClients(rnd *rand.Rand, cfgList []*struct {
	Group string
	Key   string
	Count int
}) (r []*struct {
	Consume []*struct {
		Group string
		Key   string
		Idx   int
	}
}) {
	const maxConsuming = 100
	for nnn := 0; nnn < MaxClientGenerated; nnn++ {
		clientMap := make(map[string]*struct {
			Group string
			Key   string
			Idx   int
		})

		// index 0 = 25%
		if rnd.IntN(4) == 0 {
			clientMap[IndexAsKey(cfgList, 0)] = &struct {
				Group string
				Key   string
				Idx   int
			}{Group: cfgList[0].Group, Key: cfgList[0].Key, Idx: 0}
		}
		// index 1-10 = 5%
		if rnd.IntN(20/10) == 0 {
			idx := rnd.IntN(10) + 1
			clientMap[IndexAsKey(cfgList, idx)] = &struct {
				Group string
				Key   string
				Idx   int
			}{Group: cfgList[idx].Group, Key: cfgList[idx].Key, Idx: idx}
		}
		// index 11-100 = 1%
		if rnd.IntN(100/90) == 0 {
			idx := rnd.IntN(90) + 11
			clientMap[IndexAsKey(cfgList, idx)] = &struct {
				Group string
				Key   string
				Idx   int
			}{Group: cfgList[idx].Group, Key: cfgList[idx].Key, Idx: idx}
		}
		// index 101-1000 = 0.1%
		if rnd.IntN(1000/900) == 0 {
			idx := rnd.IntN(900) + 101
			clientMap[IndexAsKey(cfgList, idx)] = &struct {
				Group string
				Key   string
				Idx   int
			}{Group: cfgList[idx].Group, Key: cfgList[idx].Key, Idx: idx}
		}
		// rest = random distribute
		for {
			if len(clientMap) >= maxConsuming {
				break
			}
			idx := rnd.IntN(len(cfgList))
			k := IndexAsKey(cfgList, idx)
			if _, ok := clientMap[k]; ok {
				// skip duplicated
				continue
			} else {
				clientMap[k] = &struct {
					Group string
					Key   string
					Idx   int
				}{Group: cfgList[idx].Group, Key: cfgList[idx].Key, Idx: idx}
			}
		}

		// strip
		elem := &struct {
			Consume []*struct {
				Group string
				Key   string
				Idx   int
			}
		}{
			Consume: make([]*struct {
				Group string
				Key   string
				Idx   int
			}, 0, maxConsuming),
		}
		for _, v := range clientMap {
			elem.Consume = append(elem.Consume, v)
		}
		r = append(r, elem)
		if len(r)%10000 == 0 {
			fmt.Println("generated clients:", len(r))
		}
	}
	return
}

func IndexAsKey(cfgList []*struct {
	Group string
	Key   string
	Count int
}, idx int) string {
	return cfgList[idx].Key + cfgList[idx].Group
}

func DataGenerateConfigureKeys(rnd *rand.Rand) []*struct {
	Group string
	Key   string
	Count int
} {
	var cfgList []*struct {
		Group string
		Key   string
		Count int
	}
	cfgkeys := map[string]struct{}{}
	for {
		cfgkeys[GenerateString(rnd, 20)] = struct{}{}
		if len(cfgkeys) >= MaxConfigurationGenerated {
			break
		} else if len(cfgkeys)%10000 == 0 {
			fmt.Println("generated configure keys:", len(cfgkeys))
		}
	}
	for k, _ := range cfgkeys {
		cfgList = append(cfgList, &struct {
			Group string
			Key   string
			Count int
		}{Group: k, Key: GenerateString(rnd, 20), Count: 0})
	}
	rnd.Shuffle(len(cfgList), func(i, j int) {
		cfgList[i], cfgList[j] = cfgList[j], cfgList[i]
	})
	return cfgList
}

func GenerateString(rnd *rand.Rand, length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rnd.IntN(len(letterRunes))]
	}
	return string(b)
}
