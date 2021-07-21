package config

// Martini contains configuration for the Martini Service.
type Martini struct {
	Emailer MartiniEmailer `json:"emailer" yaml:"emailer"`
}

// MartiniEmailer contains configuration for the Martini Emailer.
type MartiniEmailer struct {
	SenderEmail string `json:"sender_email" yaml:"sender_email"`
	SenderPwd   string `json:"sender_pwd" yaml:"sender_pwd"`

	// Used for testing to redirect all validation emails to this address
	ReceiverEmail string `json:"receiver_email" yaml:"receiver_email"`
}
