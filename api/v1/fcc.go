package v1

// FccFrnResponse provides response for https://apps.fcc.gov/cgb/form499/499results.cfm?FilerID=<fillerID>&XML=TRUE
type FccFrnResponse struct {
	FRN string `json:"frn"`
}

// FccSearchDetailResponse provides response for https://apps.fcc.gov/coresWeb/searchDetail.do?frn=<fnr>
type FccSearchDetailResponse struct {
	Email string `json:"email"`
}
