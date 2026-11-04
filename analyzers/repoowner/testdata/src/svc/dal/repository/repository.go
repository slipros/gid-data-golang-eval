package repository

type Campaign struct{}

func (*Campaign) Read()      {}
func NewCampaign() *Campaign { return &Campaign{} }

type Alias = Campaign

type History struct{}

func (*History) Read()     {}
func NewHistory() *History { return &History{} }
