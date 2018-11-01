// +build integrationtest

package integrationtest

import (
	"testing"

	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func testBasicCloudMigration(t *testing.T) {
	t.Run("sanityCloudMigrationTest", sanityCloudMigrationTest)
	t.Run("sanityMigrationTest", sanityMigrationTest)
}

func sanityCloudMigrationTest(t *testing.T) {
	// TODO: get px token information from torpedo' get storage token call
	// once it's merged in topedo , govendor and upate here
	var err error
	logrus.Info("Basic sanity tests for cloud migration")
	err = dumpRemoteKubeConfig(remoteConfig)
	require.NoError(t, err, "Error writing to clusterpair.yml: ")
	info, err := volumeDriver.GetClusterPairingInfo()
	require.NoError(t, err, "Error writing to clusterpair.yml: ")
	specReq := ClusterPairRequest{
		PairName:           remotePairName,
		ConfigMapName:      remoteConfig,
		SpecDirPath:        "./",
		RemoteIP:           info[clusterIP],
		RemoteClusterToken: info[clusterToken],
		RemotePort:         info[clusterPort],
	}
	logrus.Info("writing to spec file", specReq)
	err = CreateClusterPairSpec(specReq)
	require.NoError(t, err, "Error creating cluster Spec")
	// appply cluster pair spec and check status

	err = schedulerDriver.CreateCRDObjects("./" + pairFileName)
	require.NoError(t, err, "Error applying clusterpair")
	logrus.Info("applied to  clusterpair")
}

/* func runMysqlPods() {
	ctxs, err := schedulerDriver.Schedule(generateInstanceID(t, "singlepvctest"),
		scheduler.ScheduleOptions{AppKeys: []string{"mysql-1-pvc"}})
	require.NoError(t, err, "Error scheduling task")
	require.Equal(t, 1, len(ctxs), "Only one task should have started")

	err = schedulerDriver.WaitForRunning(ctxs[0], 600, 10)
	require.NoError(t, err, "Error waiting for pod to get to running state")
}
*/
// apply cloudmigration spec and check status
func sanityMigrationTest(t *testing.T) {
	ctxs, err := schedulerDriver.Schedule("singlemysql",
		scheduler.ScheduleOptions{AppKeys: []string{"mysql-1-pvc"}})
	require.NoError(t, err, "Error scheduling task")
	require.Equal(t, 1, len(ctxs), "Only one task should have started")

	err = schedulerDriver.WaitForRunning(ctxs[0], defaultWaitTimeout, defaultWaitInterval)
	require.NoError(t, err, "Error waiting for pod to get to running state")

	logrus.Info("Run Migration spec")
	err = schedulerDriver.CreateCRDObjects("./migrs/migration.yml")
	require.NoError(t, err, "Error applying migration specs")

	ctxs[0].KubeConfig = remoteFilePath
	err = schedulerDriver.WaitForRunning(ctxs[0], defaultWaitTimeout, defaultWaitInterval)
	require.NoError(t, err, "Error waiting for pod to get to running state")

	//	destroyAndWait(t, ctxs)

	ctxs[0].KubeConfig = ""
	err = schedulerDriver.WaitForRunning(ctxs[0], defaultWaitTimeout, defaultWaitInterval)
	require.NoError(t, err, "Error waiting for pod to get to running state")

	//	destroyAndWait(t, ctxs)
}
