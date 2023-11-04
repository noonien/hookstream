package main

import "sync"

var (
	topicsmu sync.Mutex
	topics   = map[string][]Subscriber{}
)

type Subscriber interface {
	Send([]byte)
}

// pub sends data to all subscribers of a given topic
func pub(topic string, data []byte) {
	topicsmu.Lock()
	defer topicsmu.Unlock()

	if topic, ok := topics[topic]; ok {
		for _, sub := range topic {
			sub.Send(data)
		}
	}

	if topic, ok := topics[""]; ok {
		for _, sub := range topic {
			sub.Send(data)
		}
	}
}

// sub registers a new Subscriber for a topic
func sub(topic string, sub Subscriber) {
	topicsmu.Lock()
	defer topicsmu.Unlock()

	subs, _ := topics[topic]
	topics[topic] = append(subs, sub)
}

// unsub removes a new Subscriber from a topic
func unsub(topic string, sub Subscriber) {
	topicsmu.Lock()
	defer topicsmu.Unlock()

	subs, ok := topics[topic]
	if !ok {
		return
	}

	for i, tsub := range subs {
		if tsub == sub {
			subs = append(subs[:i], subs[i+1:]...)
			break
		}
	}

	if len(subs) == 0 {
		delete(topics, topic)
	} else {
		topics[topic] = subs
	}
}
