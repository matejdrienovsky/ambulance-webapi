package ambulance_wl

import (
	"slices"
	"time"
)

func (a *Ambulance) reconcileWaitingList() {
	if len(a.WaitingList) == 0 {
		return
	}

	slices.SortFunc(a.WaitingList, func(left, right WaitingListEntry) int {
		if left.WaitingSince.Before(right.WaitingSince) {
			return -1
		}
		if left.WaitingSince.After(right.WaitingSince) {
			return 1
		}
		return 0
	})

	// We assume the first EstimatedStart is already meaningful from previous writes,
	// but enforce not being earlier than waitingSince nor now.
	if a.WaitingList[0].EstimatedStart.Before(a.WaitingList[0].WaitingSince) {
		a.WaitingList[0].EstimatedStart = a.WaitingList[0].WaitingSince
	}
	now := time.Now()
	if a.WaitingList[0].EstimatedStart.Before(now) {
		a.WaitingList[0].EstimatedStart = now
	}

	nextEntryStart := a.WaitingList[0].EstimatedStart.
		Add(time.Duration(a.WaitingList[0].EstimatedDurationMinutes) * time.Minute)

	for i := 1; i < len(a.WaitingList); i++ {
		if a.WaitingList[i].EstimatedStart.Before(nextEntryStart) {
			a.WaitingList[i].EstimatedStart = nextEntryStart
		}
		if a.WaitingList[i].EstimatedStart.Before(a.WaitingList[i].WaitingSince) {
			a.WaitingList[i].EstimatedStart = a.WaitingList[i].WaitingSince
		}

		nextEntryStart = a.WaitingList[i].EstimatedStart.
			Add(time.Duration(a.WaitingList[i].EstimatedDurationMinutes) * time.Minute)
	}
}
