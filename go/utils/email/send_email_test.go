package email

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendEmail(t *testing.T) {
	url := "https://app.market-maven-ai.com/email-verification?token="
	htmlPath := "../../assets/email_signup.html"
	placeholder := map[string]interface{}{
		"url": url,
	}

	err := SendEmail("dwiwahyudi1996@gmail.com", "Simple Auth System - Test", htmlPath, placeholder)
	require.NoError(t, err)
}
