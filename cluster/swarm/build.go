package swarm

import (
	"sync"
	"time"

	"github.com/docker/swarm/scheduler/node"
)

func newBuildSyncer() *buildSyncer {
	return &buildSyncer{
		nodeByBuildID:   map[string]*node.Node{},
		nodeBySessionID: map[string]*node.Node{},
		queueSession:    map[string]chan struct{}{},
		queueBuild:      map[string]chan struct{}{},
	}
}

type buildSyncer struct {
	mu              sync.Mutex
	nodeByBuildID   map[string]*node.Node
	nodeBySessionID map[string]*node.Node
	queueSession    map[string]chan struct{}
	queueBuild      map[string]chan struct{}
}

func (b *buildSyncer) waitSessionNode(sessionID string, timeout time.Duration) (*node.Node, error) {
	var ch chan struct{}
	b.mu.Lock()
	n, ok := b.nodeBySessionID[sessionID]
	if !ok {
		ch, ok = b.queueSession[sessionID]
		if !ok {
			ch = make(chan struct{})
		}
		b.queueSession[sessionID] = ch
	}
	b.mu.Unlock()

	select {
	case <-time.After(timeout):
		b.mu.Lock()
		delete(b.queueSession, sessionID)
		b.mu.Unlock()
		return nil, errors.Errorf("timeout waiting for build to start")
	case <-ch:
		b.mu.Lock()
		delete(b.queueSession, sessionID)
		n, ok := b.nodeBySessionID[sessionID]
		if !ok {
			return nil, errors.Errorf("build closed")
		}
		b.mu.Unlock()
	}
}

func (b *buildSyncer) waitBuildNode(buildID string, timeout time.Duration) (*node.Node, error) {
	var ch chan struct{}
	b.mu.Lock()
	n, ok := b.nodeByBuildID[buildID]
	if !ok {
		ch, ok = b.queueBuild[buildID]
		if !ok {
			ch = make(chan struct{})
		}
		b.queueBuild[buildId] = ch
	}
	b.mu.Unlock()

	select {
	case <-time.After(timeout):
		return nil, errors.Errorf("timeout waiting for build to start")
	case <-ch:
		b.mu.Lock()
		n, ok := b.nodeByBuildID[buildID]
		if !ok {
			return nil, errors.Errorf("build closed")
		}
		b.mu.Unlock()
	}
}

func (b *buildSyncer) startBuild(sessionID, buildID string, node *node.Node) (func(), error) {
	b.mu.Lock()
	_, ok := b.nodeByBuildID[buildID]
	if ok {
		return nil, errors.Errorf("build already started")
	}
	b.nodeByBuildID[buildID] = node
	b.nodeBySessionID[sessionID] = node
	if ch, ok := b.queueBuild[buildID]; ok {
		close(ch)
		delete(b.queueBuild, buildID)
	}
	if ch, ok := b.queueSession[sessionID]; ok {
		close(ch)
		delete(b.queueSession, sessionID)
	}
	b.mu.Unlock()
	return func() {
		b.mu.Lock()
		delete(b.nodeBySessionID, sessionID)
		delete(b.nodeByBuildID, buildID)
		b.mu.Unlock()
	}, nil
}
