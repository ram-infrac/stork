// +build integrationtest

package integrationtest

import (
	"testing"
	"time"

	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func testBasicCloudMigration(t *testing.T) {
	t.Run("sanityCloudMigrationTest", sanityCloudMigrationTest)
	t.Run("sanityMigrationTest", sanityMigrationTest)
}
func sanityCloudMigrationTest(t *testing.T) {
	var err error
	logrus.Info("Basic sanity tests for cloud migration")

	err = dumpRemoteKubeConfig(remoteConfig)
	require.NoError(t, err, "Error writing to clusterpair.yml: ")

	info, err := volumeDriver.GetClusterPairingInfo()
	require.NoError(t, err, "Error writing to clusterpair.yml: ")

	specReq := ClusterPairRequest{
		PairName:           remotePairName,
		ConfigMapName:      remoteConfig,
		SpecDirPath:        "./migrs/",
		RemoteIP:           info[clusterIP],
		RemoteClusterToken: info[clusterToken],
		RemotePort:         info[clusterPort],
	}
	logrus.Info("Writing to spec file", specReq)
	err = CreateClusterPairSpec(specReq)
	require.NoError(t, err, "Error creating cluster Spec")

	ctx, err := getContextCRD("cluster-pair")
	require.NoError(t, err, "Error locating cluster Spec")

	err = schedulerDriver.CreateCRDObjects(ctx, 2*time.Minute, 10*time.Second)
	require.NoError(t, err, "Error applying clusterpair")

	logrus.Info("Applied to  clusterpair")
}

// apply cloudmigration spec and check status
func sanityMigrationTest(t *testing.T) {
	ctxs, err := schedulerDriver.Schedule("singlemysql",
		scheduler.ScheduleOptions{AppKeys: []string{"mysql-1-pvc"}})
	require.NoError(t, err, "Error scheduling task")
	require.Equal(t, 1, len(ctxs), "Only one task should have started")
	err = schedulerDriver.WaitForRunning(ctxs[0], defaultWaitTimeout, defaultWaitInterval)
	require.NoError(t, err, "Error waiting for pod to get to running state")
	logrus.Info("Run Migration spec")

	// appply cluster pair spec and check status
	ctx, err := getContextCRD("migration")
	require.NoError(t, err, "Error locating migration Spec")

	err = schedulerDriver.CreateCRDObjects(ctx, 2*time.Minute, 10*time.Second)
	require.NoError(t, err, "Error applying migration specs")

	err = schedulerDriver.WaitForRunning(ctxs[0], defaultWaitTimeout, defaultWaitInterval)
	require.NoError(t, err, "Error waiting for pod to get to running state")
	//	destroyAndWait(t, ctxs)
	err = schedulerDriver.WaitForRunning(ctxs[0], defaultWaitTimeout, defaultWaitInterval)
	require.NoError(t, err, "Error waiting for pod to get to running state")
	//	destroyAndWait(t, ctxs)
}
