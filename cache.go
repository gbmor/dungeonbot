package main

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"sync"
)

type notesCache struct {
	sync.RWMutex
	kv map[string]string
	pb string
}

func pastebin(pastebin string, input string) (string, error) {
	pbconn, err := net.Dial("tcp", pastebin)
	if err != nil {
		return "", fmt.Errorf("Error connecting to pastebin service: %w", err)
	}
	defer pbconn.Close()

	if _, err := pbconn.Write([]byte(input)); err != nil {
		return "", fmt.Errorf("Error sending data to pastebin service: %w", err)
	}

	pbRdr := bufio.NewReader(pbconn)
	pbBytes, _, err := pbRdr.ReadLine()
	if err != nil {
		return "", fmt.Errorf("Error reading response from pastebin service: %w", err)
	}

	return string(pbBytes), err
}

func (c *notesCache) bap(notes string) string {
	return c.yoink(c.yeet(notes))
}

func (c *notesCache) yoink(hash string) string {
	c.RLock()
	defer c.RUnlock()

	if out, ok := c.kv[hash]; ok {
		return out
	}

	return "not in cache"
}

func (c *notesCache) yeet(notes string) string {
	if notes == "" {
		return "empty notes provided"
	}

	bytes := fnv.New32a().Sum([]byte(notes))
	hash := fmt.Sprintf("%x", bytes)

	c.Lock()
	defer c.Unlock()

	if _, ok := c.kv[hash]; ok {
		return hash
	}

	url, err := pastebin(c.pb, notes)
	if err != nil {
		log.Printf("cache.yeet(): %s", err.Error())
		return "pastebin error"
	}

	c.kv[hash] = url
	return hash
}
