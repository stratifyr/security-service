package stores

import (
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http"
)

type IndustryStore interface {
	Index(ctx *gofr.Context) []Industry
}

const (
	AutomobileAndAutoComponents Industry = iota
	CapitalGoods
	Chemicals
	Construction
	ConstructionMaterials
	ConsumerDurables
	ConsumerServices
	Diversified
	FastMovingConsumerGoods
	FinancialServices
	ForestMaterials
	Healthcare
	InformationTechnology
	MediaEntertainmentAndPublication
	MetalsAndMining
	OilGasAndConsumableFuels
	Power
	Realty
	Services
	Telecommunication
	Textiles
	Index
	Bond
	Gold
	Silver
	Utilities
)

type Industry int

type industryStore struct{}

func NewIndustryStore() IndustryStore {
	return &industryStore{}
}

func (s *industryStore) Index(ctx *gofr.Context) []Industry {
	return []Industry{
		AutomobileAndAutoComponents,
		CapitalGoods,
		Chemicals,
		Construction,
		ConstructionMaterials,
		ConsumerDurables,
		ConsumerServices,
		Diversified,
		FastMovingConsumerGoods,
		FinancialServices,
		ForestMaterials,
		Healthcare,
		InformationTechnology,
		MediaEntertainmentAndPublication,
		MetalsAndMining,
		OilGasAndConsumableFuels,
		Power,
		Realty,
		Services,
		Telecommunication,
		Textiles,
		Index,
		Bond,
		Gold,
		Silver,
		Utilities,
	}
}

func (ex Industry) String() string {
	var conversionMap = map[Industry]string{
		AutomobileAndAutoComponents:      "Automobile and Auto Components",
		CapitalGoods:                     "Capital Goods",
		Chemicals:                        "Chemicals",
		Construction:                     "Construction",
		ConstructionMaterials:            "Construction Materials",
		ConsumerDurables:                 "Consumer Durables",
		ConsumerServices:                 "Consumer Services",
		Diversified:                      "Diversified",
		FastMovingConsumerGoods:          "Fast Moving Consumer Goods",
		FinancialServices:                "Financial Services",
		ForestMaterials:                  "Forest Materials",
		Healthcare:                       "Healthcare",
		InformationTechnology:            "Information Technology",
		MediaEntertainmentAndPublication: "Media Entertainment & Publication",
		MetalsAndMining:                  "Metals & Mining",
		OilGasAndConsumableFuels:         "Oil Gas & Consumable Fuels",
		Power:                            "Power",
		Realty:                           "Realty",
		Services:                         "Services",
		Telecommunication:                "Telecommunication",
		Textiles:                         "Textiles",
		Index:                            "Index",
		Bond:                             "Bond",
		Gold:                             "Gold",
		Silver:                           "Silver",
		Utilities:                        "Utilities",
	}

	return conversionMap[ex]
}

func IndustryFromString(str string) (Industry, error) {
	var conversionMap = map[string]Industry{
		"Automobile and Auto Components":    AutomobileAndAutoComponents,
		"Capital Goods":                     CapitalGoods,
		"Chemicals":                         Chemicals,
		"Construction":                      Construction,
		"Construction Materials":            ConstructionMaterials,
		"Consumer Durables":                 ConsumerDurables,
		"Consumer Services":                 ConsumerServices,
		"Diversified":                       Diversified,
		"Fast Moving Consumer Goods":        FastMovingConsumerGoods,
		"Financial Services":                FinancialServices,
		"Forest Materials":                  ForestMaterials,
		"Healthcare":                        Healthcare,
		"Information Technology":            InformationTechnology,
		"Media Entertainment & Publication": MediaEntertainmentAndPublication,
		"Metals & Mining":                   MetalsAndMining,
		"Oil Gas & Consumable Fuels":        OilGasAndConsumableFuels,
		"Power":                             Power,
		"Realty":                            Realty,
		"Services":                          Services,
		"Telecommunication":                 Telecommunication,
		"Textiles":                          Textiles,
		"Index":                             Index,
		"Bond":                              Bond,
		"Gold":                              Gold,
		"Silver":                            Silver,
		"Utilities":                         Utilities,
	}

	industry, ok := conversionMap[str]
	if !ok {
		return 0, http.ErrorEntityNotFound{Name: "industries", Value: str}
	}

	return industry, nil
}
