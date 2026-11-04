package service

type Reader interface{ Read() }
type Campaign struct{ repo Reader }
type Export struct{ Repo Reader }
type History struct{ repo Reader }

func NewCampaign(repo Reader) *Campaign        { return &Campaign{repo: repo} }
func NewExport(repo Reader) *Export            { return &Export{Repo: repo} }
func NewHistory(repo Reader) (*History, error) { return &History{repo: repo}, nil }

type CampaignOptions struct{ Repo Reader }
