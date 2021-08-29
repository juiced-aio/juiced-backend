package staticstores

func InitStores() error {
	var err error

	err = userInfoStore.Init()
	if err != nil {
		return err
	}

	err = settingsStore.Init()
	if err != nil {
		return err
	}

	err = checkoutsStore.Init()
	if err != nil {
		return err
	}

	return nil
}
