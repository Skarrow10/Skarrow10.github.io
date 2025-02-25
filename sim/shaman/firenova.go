package shaman

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func (shaman *Shaman) registerFireNovaSpell() {
	manaCost := 0.22 * shaman.BaseMana

	fireNovaGlyphCDReduction := core.TernaryInt32(shaman.HasMajorGlyph(proto.ShamanMajorGlyph_GlyphOfFireNova), 3, 0)
	impFireNovaCDReduction := shaman.Talents.ImprovedFireNova * 2
	fireNovaCooldown := 10 - fireNovaGlyphCDReduction - impFireNovaCDReduction

	shaman.FireNova = shaman.RegisterSpell(core.SpellConfig{
		ActionID:     core.ActionID{SpellID: 61657},
		SpellSchool:  core.SpellSchoolFire,
		ProcMask:     core.ProcMaskSpellDamage,
		Flags:        SpellFlagFocusable,
		ResourceType: stats.Mana,
		BaseCost:     manaCost,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost: manaCost,
				GCD:  core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    shaman.NewTimer(),
				Duration: time.Second * time.Duration(fireNovaCooldown),
			},
			ModifyCast: func(_ *core.Simulation, spell *core.Spell, cast *core.Cast) {
				shaman.modifyCastClearcasting(spell, cast)
			},
		},

		BonusHitRating:   float64(shaman.Talents.ElementalPrecision) * core.SpellHitRatingPerHitChance,
		DamageMultiplier: 1 + float64(shaman.Talents.CallOfFlame)*0.05 + float64(shaman.Talents.ImprovedFireNova)*0.1,
		CritMultiplier:   shaman.ElementalCritMultiplier(0),
		ThreatMultiplier: 1 - (0.1/3)*float64(shaman.Talents.ElementalPrecision),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// FIXME: double check spell coefficients
			dmgFromSP := 0.2142 * spell.SpellPower()
			for _, aoeTarget := range sim.Encounter.Targets {
				baseDamage := sim.Roll(893, 997) + dmgFromSP
				// TODO: Uncomment this
				//baseDamage *= sim.Encounter.AOECapMultiplier()
				spell.CalcAndDealDamage(sim, &aoeTarget.Unit, baseDamage, spell.OutcomeMagicHitAndCrit)
			}
		},
	})
}

func (shaman *Shaman) IsFireNovaCastable(sim *core.Simulation) bool {
	return shaman.FireNova.IsReady(sim) && shaman.Totems.Fire > 0
}
