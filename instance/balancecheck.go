// Copyright (C) 2021 The Dank Grinder authors.
//
// This source code has been released under the GNU Affero General Public
// License v3.0. A copy of this license is available at
// https://www.gnu.org/licenses/agpl-3.0.en.html

package instance

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dankgrinder/dankgrinder/instance/scheduler"

	"github.com/dankgrinder/dankgrinder/discord"
)

func (in *Instance) balanceCheck(msg discord.Message) {
	if !strings.Contains(msg.Embeds[0].Title, in.Client.User.Username) {
		return
	}
	if !exp.bal.Match([]byte(msg.Embeds[0].Description)) {
		return
	}

	match := exp.bal.FindStringSubmatch(msg.Embeds[0].Description)

	balstr := strings.Replace(match[1], ",", "", -1)
	balance, err := strconv.Atoi(balstr)
	if err != nil {
		in.Logger.Errorf("error while reading balance: %v", err)
		return
	}
	
	netstr := strings.Replace(match[2], ",", "", -1)
	networth, err := strconv.Atoi(netstr)
	if err != nil {
		in.Logger.Errorf("error while reading net worth: %v", err)
		return
	}
	
	in.updateBalance(balance, networth)
}

func (in *Instance) updateBalance(balance, networth int) {
	if balance > in.Features.AutoShare.MaximumBalance &&
		in.Features.AutoShare.Enable &&
		in.Master != nil &&
		in != in.Master {
		in.sdlr.PrioritySchedule(&scheduler.Command{
			Value: fmt.Sprintf(
				"pls trade %v <@%v>",
				balance-in.Features.AutoShare.MinimumBalance,
				in.Master.Client.User.ID,
			),
			Log: "sharing all balance above minimum with master instance",
			AwaitResume: true,
		})
	}
	in.balance = balance
	in.lastBalanceUpdate = time.Now()
	in.Logger.Infof(
		"current wallet balance: %v coins, current net worth: %v coins",
		numFmt.Sprintf("%d", balance),
		numFmt.Sprintf("%d", networth),
	)

	

	if in.startingTime.IsZero() {
		in.initialBalance = balance
		netWorth[in.Client.User.Username] = networth
		in.startingTime = time.Now()
		return
	}

	inc := balance - in.initialBalance
	netInc := networth - netWorth[in.Client.User.Username]
	per := time.Since(in.startingTime)
	hourlyInc := int(math.Round(float64(inc) / per.Hours()))
	hourlyNetInc := int(math.Round(float64(netInc) / per.Hours()))
	in.Logger.Infof(
		"average income: %v coins/h, %v net worth/h",
		numFmt.Sprintf("%d", hourlyInc),
		numFmt.Sprintf("%d", hourlyNetInc),
	)
	
	// update networth dictionary
	netWorth[in.Client.User.Username] = networth
}
