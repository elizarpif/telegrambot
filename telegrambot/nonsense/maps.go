package nonsense

import (
	"sync"

	"github.com/google/uuid"
)

type usersInGameNap struct {
	syncMap sync.Map
}

func newUsersInGameNap() *usersInGameNap {
	return &usersInGameNap{}
}

func (u *usersInGameNap) Load(uid uuid.UUID) (*game, bool) {
	ggame, ok := u.syncMap.Load(uid)
	if !ok {
		return nil, ok
	}

	game, _ := ggame.(*game)
	return game, ok
}

func (u *usersInGameNap) Store(uid uuid.UUID, g *game) {
	u.syncMap.Store(uid, g)
}

func (u *usersInGameNap) Delete(uid uuid.UUID) {
	u.syncMap.Delete(uid)
}

// users        map[int64]uuid.UUID
type usersMap struct {
	syncMap sync.Map
}

func newUsersMap() *usersMap {
	return &usersMap{}
}

func (u *usersMap) Load(id int64) (uuid.UUID, bool) {
	user, ok := u.syncMap.Load(id)
	if !ok {
		return uuid.UUID{}, ok
	}
	userRes, _ := user.(uuid.UUID)
	return userRes, ok
}

func (u *usersMap) Store(id int64, uid uuid.UUID) {
	u.syncMap.Store(id, uid)
}

func (u *usersMap) Delete(id int64) {
	u.syncMap.Delete(id)
}

type waitUidUsers struct {
	// map[int64]struct{}
	syncMap sync.Map
}

func newWaitUidUsers() *waitUidUsers {
	return &waitUidUsers{}
}

func (u *waitUidUsers) Load(id int64) (uuid.UUID, bool) {
	user, ok := u.syncMap.Load(id)
	if !ok {
		return uuid.UUID{}, ok
	}
	userRes, _ := user.(uuid.UUID)
	return userRes, ok
}

func (u *waitUidUsers) Store(id int64) {
	u.syncMap.Store(id, struct{}{})
}

func (u *waitUidUsers) Delete(id int64) {
	u.syncMap.Delete(id)
}
