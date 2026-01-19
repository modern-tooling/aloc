package aggregator

import "github.com/rgehrsern/aloc/internal/model"

func ComputeRatios(responsibilities []model.Responsibility) model.Ratios {
	byRole := make(map[model.Role]int)
	for _, r := range responsibilities {
		byRole[r.Role] = r.LOC
	}

	prodLOC := byRole[model.RoleProd]
	if prodLOC == 0 {
		prodLOC = 1 // avoid division by zero
	}

	return model.Ratios{
		TestToProd:      float32(byRole[model.RoleTest]) / float32(prodLOC),
		InfraToProd:     float32(byRole[model.RoleInfra]) / float32(prodLOC),
		DocsToProd:      float32(byRole[model.RoleDocs]) / float32(prodLOC),
		GeneratedToProd: float32(byRole[model.RoleGenerated]) / float32(prodLOC),
		ConfigToProd:    float32(byRole[model.RoleConfig]) / float32(prodLOC),
	}
}
