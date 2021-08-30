package sound

func DeleteSoundWithGroups(groupProvider *GroupProvider, soundProvider *SoundProvider, sound *Sound) (err error) {
	var deleteGroups []*Group

	for _, group := range groupProvider.List() {
		deleteGroups = append(deleteGroups, group)
	}

	err = groupProvider.Delete(deleteGroups...)
	if err != nil {
		return err
	}

	return soundProvider.Delete(sound)
}
