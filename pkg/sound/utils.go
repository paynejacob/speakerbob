package sound

func DeleteSound(groupProvider *GroupProvider, soundProvider *SoundProvider, s *Sound) error {
	groups := make([]*Group, 0)
	groupKeys := make([]string, 0)

	for _, group := range groupProvider.List() {
		if contains(group.SoundIds, s.Id) {
			groups = append(groups, group)
			groupKeys = append(groupKeys, groupProvider.GetKey(group))
		}
	}

	if err := soundProvider.Store.Delete(append(groupKeys, soundProvider.GetKey(s), soundProvider.AudioKey(s))...); err != nil {
		return err
	}

	if err := soundProvider.Delete(s); err != nil {
		return err
	}

	for i := 0; i < len(groups); i++ {
		if err := groupProvider.Delete(groups[i]); err != nil {
			return err
		}
	}

	return nil
}

func contains(arr []string, v string) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == v {
			return true
		}
	}

	return false
}
