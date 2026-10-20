// Package campaign holds the action that campaigns in an election:
//
//	err := anActor.AttemptsTo(campaign.InTheElection("primary-worker"))
//	err := anActor.AttemptsTo(campaign.InTheElection("primary-worker").WithTheValue("node-1"))
package campaign
