package sound

func DeleteSoundWithGroups(groupProvider *GroupProvider, soundProvider *SoundProvider, sound *Sound) (err error) {
	var deleteGroups []*Group

	for _, group := range groupProvider.List() {
		for _, soundId := range group.SoundIds {
			if soundId == sound.Id {
				deleteGroups = append(deleteGroups, group)
				break
			}
		}
	}

	if deleteGroups != nil {
		err = groupProvider.Delete(deleteGroups...)
		if err != nil {
			return err
		}
	}

	return soundProvider.Delete(sound)
}
