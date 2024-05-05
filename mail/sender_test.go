package mail

import (
	"testing"

	"github.com/dubass83/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmail(t *testing.T) {
	// if testing.Short() {
	// 	t.Skip()
	// }
	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewMailtrapSender(
		config.EmailSenderName,
		config.EmailSenderEmailFrom,
		config.MailtrapLogin,
		config.MailtrapPass,
	)

	subject := "Hello from Simple Bank!"
	content := `<h1>Hi there!</h1></br>
	This is test message from simple bank dev project</br>
	<p>You can find source code <a href="https://github.com/dubass83/simple_bank_golang">here</a></p>`
	to := []string{"makssych@gmail.com", "makssych@outlook.com"}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
