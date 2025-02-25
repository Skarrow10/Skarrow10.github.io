package paladin

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func (paladin *Paladin) registerExorcismSpell() {
	// From the perspective of max rank.
	baseCost := paladin.BaseMana * 0.08 * (1 - 0.02*float64(paladin.Talents.Benediction))

	paladin.Exorcism = paladin.RegisterSpell(core.SpellConfig{
		ActionID:     core.ActionID{SpellID: 48801},
		SpellSchool:  core.SpellSchoolHoly,
		ProcMask:     core.ProcMaskSpellDamage,
		ResourceType: stats.Mana,
		BaseCost:     baseCost,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost:     baseCost,
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * 1500,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Second * 15,
			},
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				if paladin.ArtOfWarInstantCast.IsActive() {
					paladin.ArtOfWarInstantCast.Deactivate(sim)
					cast.CastTime = 0
					return
				}
				paladin.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+cast.CastTime)
			},
		},

		DamageMultiplierAdditive: 1 +
			paladin.getTalentSanctityOfBattleBonus() +
			paladin.getMajorGlyphOfExorcismBonus() +
			paladin.getItemSetAegisBattlegearBonus2(),
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   paladin.SpellCritMultiplier(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(1028, 1146) +
				.15*spell.SpellPower() +
				.15*spell.MeleeAttackPower()

			isDemonOrUndead := target.MobType == proto.MobType_MobTypeDemon || target.MobType == proto.MobType_MobTypeUndead
			if isDemonOrUndead {
				spell.BonusCritRating += 100 * core.CritRatingPerCritChance
			}
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			if isDemonOrUndead {
				spell.BonusCritRating -= 100 * core.CritRatingPerCritChance
			}
		},
	})
}
