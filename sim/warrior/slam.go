package warrior

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func (warrior *Warrior) registerSlamSpell() {
	cost := 15.0 - float64(warrior.Talents.FocusedRage)
	refundAmount := cost * 0.8

	warrior.Slam = warrior.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 47475},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagIncludeTargetBonusDamage,

		ResourceType: stats.Rage,
		BaseCost:     cost,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost:     cost,
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond*1500 - time.Millisecond*500*time.Duration(warrior.Talents.ImprovedSlam),
			},
			IgnoreHaste: true,
		},

		BonusCritRating:  core.TernaryFloat64(warrior.HasSetBonus(ItemSetWrynnsBattlegear, 4), 5, 0) * core.CritRatingPerCritChance,
		DamageMultiplier: 1 + 0.02*float64(warrior.Talents.UnendingFury) + core.TernaryFloat64(warrior.HasSetBonus(ItemSetDreadnaughtBattlegear, 2), 0.1, 0),
		CritMultiplier:   warrior.critMultiplier(mh),
		ThreatMultiplier: 1,
		FlatThreatBonus:  140,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := 250 +
				spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower()) +
				spell.BonusWeaponDamage()

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if !result.Landed() {
				warrior.AddRage(sim, refundAmount, warrior.RageRefundMetrics)
			}
		},
	})
}

func (warrior *Warrior) ShouldInstantSlam(sim *core.Simulation) bool {
	return warrior.CurrentRage() >= warrior.Slam.DefaultCast.Cost && warrior.Slam.IsReady(sim) && warrior.isBloodsurgeActive() &&
		sim.CurrentTime > (warrior.lastBloodsurgeProc+warrior.reactionTime) && warrior.GCD.IsReady(sim)
}

func (warrior *Warrior) ShouldSlam(sim *core.Simulation) bool {
	return warrior.CurrentRage() >= warrior.Slam.DefaultCast.Cost && warrior.Slam.IsReady(sim) && warrior.Talents.ImprovedSlam > 0
}

func (warrior *Warrior) CastSlam(sim *core.Simulation, target *core.Unit) bool {
	if !warrior.Slam.Cast(sim, target) {
		return false
	}
	warrior.AutoAttacks.DelayMeleeBy(sim, warrior.Slam.CurCast.CastTime)
	warrior.disableHsCleaveUntil = sim.CurrentTime + warrior.Slam.CurCast.CastTime
	return true
}
