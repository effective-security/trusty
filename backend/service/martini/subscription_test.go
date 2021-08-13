package martini

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_subscriptionExpiryPeriodFromProductName(t *testing.T) {
	require.Equal(t, 1, subscriptionExpiryPeriodFromProductName("1 year subsription"))
	require.Equal(t, 1, subscriptionExpiryPeriodFromProductName("1 YEAR Subscription"))
	require.Equal(t, 2, subscriptionExpiryPeriodFromProductName("2 years subsription"))
	require.Equal(t, 2, subscriptionExpiryPeriodFromProductName("Subscription for 2 years"))
	require.Equal(t, 3, subscriptionExpiryPeriodFromProductName("3 years subsription"))
	require.Equal(t, 4, subscriptionExpiryPeriodFromProductName("4 years subsription"))
	require.Equal(t, 5, subscriptionExpiryPeriodFromProductName("5 years subsription"))
	require.Equal(t, 1, subscriptionExpiryPeriodFromProductName("6 years subsription"))
}
