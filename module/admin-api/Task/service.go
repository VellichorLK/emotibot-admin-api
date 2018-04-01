package Task

func EnableAllScenario(appid string) {
	setAllScenarioStatus(appid, true)
}

func DisableAllScenario(appid string) {
	setAllScenarioStatus(appid, false)
}

func EnableScenario(appid string, scenarioID string) {
	setScenarioStatus(appid, scenarioID, true)
}

func DisableScenario(appid string, scenarioID string) {
	setScenarioStatus(appid, scenarioID, false)
}
