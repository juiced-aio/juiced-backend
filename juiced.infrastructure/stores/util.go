package stores

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

	err = taskGroupStore.Init()
	if err != nil {
		return err
	}

	err = taskStore.Init()
	if err != nil {
		return err
	}

	err = profileGroupStore.Init()
	if err != nil {
		return err
	}

	err = profileStore.Init()
	if err != nil {
		return err
	}

	err = proxyGroupStore.Init()
	if err != nil {
		return err
	}

	return nil
}
