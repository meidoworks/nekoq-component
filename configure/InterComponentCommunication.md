# Inter-component communication

## Overall

Components will always have the request on working with other components under the same component suite.

This document demonstrates the ways of inter-component communication.

## Methods

Cross multiple components

* General: [component -> client -> server -> component] via standard builtin clients
* All in one: [component -> component] via integrating all components
* On-demand: [component -> component api -> client based component impl -> server -> component]

Scaling the same component

* Nested: [component -> component api -> client based component impl -> server -> component]
* Aside: [[requester 1st req -> component], [requester 2nd req -> component], ...]

