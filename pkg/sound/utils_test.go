package sound

import (
	"github.com/paynejacob/hotcereal/pkg/stores/memory"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteSoundWithGroups(t *testing.T) {
	sp := &SoundProvider{Store: memory.New()}
	_ = sp.Initialize()

	gp := &GroupProvider{Store: memory.New()}
	_ = gp.Initialize()

	s1 := NewSound()
	s1.Name = "s1"
	s2 := NewSound()
	s2.Name = "s2"
	_ = sp.Save(&s1)
	_ = sp.Save(&s2)

	g1 := NewGroup()
	g1.Name = "g1"
	g1.SoundIds = []string{s1.Id, s2.Id}
	g2 := NewGroup()
	g2.Name = "g2"
	g2.SoundIds = []string{s2.Id}
	g3 := NewGroup()
	g3.Name = "g3"
	g3.SoundIds = []string{s2.Id}
	_ = gp.Save(&g1)
	_ = gp.Save(&g2)
	_ = gp.Save(&g3)

	deletedGroups, err := DeleteSoundWithGroups(gp, sp, &s1)

	assert.NoError(t, err)
	assert.Len(t, deletedGroups, 1)
	assert.Equal(t, g1.Id, deletedGroups[0].Id)

	assert.Len(t, sp.List(), 1)
	assert.Len(t, gp.List(), 2)
}

func TestDeleteSoundWithGroupsNoCascade(t *testing.T) {
	sp := &SoundProvider{Store: memory.New()}
	_ = sp.Initialize()

	gp := &GroupProvider{Store: memory.New()}
	_ = gp.Initialize()

	s1 := NewSound()
	s1.Name = "s1"
	_ = sp.Save(&s1)

	deletedGroups, err := DeleteSoundWithGroups(gp, sp, &s1)

	assert.NoError(t, err)
	assert.Len(t, deletedGroups, 0)
	assert.Len(t, sp.List(), 0)
}
