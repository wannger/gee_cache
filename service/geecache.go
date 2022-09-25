package service

import (
	"fmt"
	"log"
	"sync"
)

// 负责与用户的交互的文件

// 回调函数接口

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 一个Group可以认为是一个缓存的命名空间，每个Group拥有一个唯一的名称name

// 传参为缓存内存大小，回调函数不能为空

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// 读缓存

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

// Get方法的实现

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 如果命中了缓存
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	// 没命中缓存，调用回调函数
	return g.load(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// 从回调函数获取值，加入缓存中
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 添加缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 将PeerPicker接口的HTTPPool注入到Group中

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from perr", err)
		}
	}
	return g.getLocally(key)
}

// 实现了PeerGetter接口的HttpGetter访问远程节点，获取缓存值

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
