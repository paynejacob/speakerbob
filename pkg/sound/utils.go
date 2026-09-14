package sound

func DeleteSoundWithGroups(groupProvider *GroupProvider, soundProvider *SoundProvider, sound *Sound) (deletedGroups []*Group, err error) {
	for _, group := range groupProvider.List() {
		for _, soundId := range group.SoundIds {
			if soundId == sound.Id {
				deletedGroups = append(deletedGroups, group)
				break
			}
		}
	}

	if deletedGroups != nil {
		if err = groupProvider.Delete(deletedGroups...); err != nil {
			return nil, err
		}
	}

	if err = soundProvider.Delete(sound); err != nil {
		return nil, err
	}

	return deletedGroups, nil
}
