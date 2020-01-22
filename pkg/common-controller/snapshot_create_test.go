/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common_controller

import (
	"errors"
	"testing"
	"time"

	crdv1 "github.com/kubernetes-csi/external-snapshotter/v2/pkg/apis/volumesnapshot/v1beta1"
	v1 "k8s.io/api/core/v1"
	storage "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var timeNow = time.Now()
var timeNowStamp = timeNow.UnixNano()
var False = false
var True = true

var metaTimeNowUnix = &metav1.Time{
	Time: timeNow,
}

var defaultSize int64 = 1000
var deletePolicy = crdv1.VolumeSnapshotContentDelete
var retainPolicy = crdv1.VolumeSnapshotContentRetain
var sameDriverStorageClass = &storage.StorageClass{
	TypeMeta: metav1.TypeMeta{
		Kind: "StorageClass",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: sameDriver,
	},
	Provisioner: mockDriverName,
	Parameters:  class1Parameters,
}

var diffDriverStorageClass = &storage.StorageClass{
	TypeMeta: metav1.TypeMeta{
		Kind: "StorageClass",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: diffDriver,
	},
	Provisioner: mockDriverName,
	Parameters:  class1Parameters,
}

// Test single call to SyncSnapshot, expecting create snapshot to happen.
// 1. Fill in the controller with initial data
// 2. Call the SyncSnapshot *once*.
// 3. Compare resulting contents with expected contents.
func TestCreateSnapshotSync(t *testing.T) {
	tests := []controllerTest{
		{
			name:              "6-1 - successful create snapshot with snapshot class gold",
			initialContents:   nocontents,
			expectedContents:  newContentArrayNoStatus("snapcontent-snapuid6-1", "snapuid6-1", "snap6-1", "sid6-1", classGold, "", "pv-handle6-1", deletionPolicy, nil, nil, false, false),
			initialSnapshots:  newSnapshotArray("snap6-1", "snapuid6-1", "claim6-1", "", classGold, "", &False, nil, nil, nil, false, true),
			expectedSnapshots: newSnapshotArray("snap6-1", "snapuid6-1", "claim6-1", "", classGold, "snapcontent-snapuid6-1", &False, nil, nil, nil, false, true),
			initialClaims:     newClaimArray("claim6-1", "pvc-uid6-1", "1Gi", "volume6-1", v1.ClaimBound, &classEmpty),
			initialVolumes:    newVolumeArray("volume6-1", "pv-uid6-1", "pv-handle6-1", "1Gi", "pvc-uid6-1", "claim6-1", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
			errors:            noerrors,
			test:              testSyncSnapshot,
		},
		{
			name:              "6-2 - successful create snapshot with snapshot class silver",
			initialContents:   nocontents,
			expectedContents:  newContentArrayNoStatus("snapcontent-snapuid6-2", "snapuid6-2", "snap6-2", "sid6-2", classSilver, "", "pv-handle6-2", deletionPolicy, nil, nil, false, false),
			initialSnapshots:  newSnapshotArray("snap6-2", "snapuid6-2", "claim6-2", "", classSilver, "", &False, nil, nil, nil, false, true),
			expectedSnapshots: newSnapshotArray("snap6-2", "snapuid6-2", "claim6-2", "", classSilver, "snapcontent-snapuid6-2", &False, nil, nil, nil, false, true),
			initialClaims:     newClaimArray("claim6-2", "pvc-uid6-2", "1Gi", "volume6-2", v1.ClaimBound, &classEmpty),
			initialVolumes:    newVolumeArray("volume6-2", "pv-uid6-2", "pv-handle6-2", "1Gi", "pvc-uid6-2", "claim6-2", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
			errors:            noerrors,
			test:              testSyncSnapshot,
		},
		{
			name:              "7-1 - fail to create snapshot with non-existing snapshot class",
			initialContents:   nocontents,
			expectedContents:  nocontents,
			initialSnapshots:  newSnapshotArray("snap7-1", "snapuid7-1", "claim7-1", "", classNonExisting, "", &False, nil, nil, nil, false, true),
			expectedSnapshots: newSnapshotArray("snap7-1", "snapuid7-1", "claim7-1", "", classNonExisting, "", &False, nil, nil, newVolumeError("Failed to create snapshot content with error failed to get input parameters to create snapshot snap7-1: \"failed to retrieve snapshot class non-existing from the informer: \\\"volumesnapshotclass.snapshot.storage.k8s.io \\\\\\\"non-existing\\\\\\\" not found\\\"\""), false, true),
			initialClaims:     newClaimArray("claim7-1", "pvc-uid7-1", "1Gi", "volume7-1", v1.ClaimBound, &classEmpty),
			initialVolumes:    newVolumeArray("volume7-1", "pv-uid7-1", "pv-handle7-1", "1Gi", "pvc-uid7-1", "claim7-1", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
			expectedEvents:    []string{"Warning SnapshotContentCreationFailed"},
			errors:            noerrors,
			expectSuccess:     false,
			test:              testSyncSnapshot,
		},

		{
			name:              "7-2 - fail to update snapshot reports warning event",
			initialContents:   newContentArrayWithReadyToUse("snapcontent-snapuid7-2", "snapuid7-2", "snap7-2", "sid7-2", classGold, "sid7-2", "pv-handle7-2", deletionPolicy, nil, nil, &True, false),
			expectedContents:  newContentArrayWithReadyToUse("snapcontent-snapuid7-2", "snapuid7-2", "snap7-2", "sid7-2", classGold, "sid7-2", "pv-handle7-2", deletionPolicy, nil, nil, &True, false),
			initialSnapshots:  newSnapshotArray("snap7-2", "snapuid7-2", "claim7-2", "", classGold, "snapcontent-snapuid7-2", &False, nil, nil, nil, false, true),
			expectedSnapshots: newSnapshotArray("snap7-2", "snapuid7-2", "claim7-2", "", classGold, "snapcontent-snapuid7-2", &False, nil, nil, newVolumeError("Snapshot status update failed, snapshot controller failed to update default/snap7-2 on API server: mock update error"), false, true),
			initialClaims:     newClaimArray("claim7-2", "pvc-uid7-2", "1Gi", "volume7-2", v1.ClaimBound, &classGold),
			initialVolumes:    newVolumeArray("volume7-2", "pv-uid7-2", "pv-handle7-2", "1Gi", "pvc-uid7-2", "claim7-2", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classGold),
			expectedEvents:    []string{"Warning SnapshotStatusUpdateFailed"},
			errors: []reactorError{
				// Inject error to the forth client.VolumesnapshotV1beta1().VolumeSnapshots().Update call.
				// All other calls will succeed.
				{"update", "volumesnapshots", errors.New("mock update error")},
				{"update", "volumesnapshots", errors.New("mock update error")},
				{"update", "volumesnapshots", errors.New("mock update error")},
			}, test: testSyncSnapshot,
		},
		/*{
			name:                  "7-3 - fail to create snapshot without snapshot class ",
			initialContents:       nocontents,
			expectedContents:      nocontents,
			initialSnapshots:      newSnapshotArray("snap7-3", "snapuid7-3", "claim7-3", "", "", "", &False, nil, nil, nil, false, true),
			expectedSnapshots:     newSnapshotArray("snap7-3", "snapuid7-3", "claim7-3", "", "", "", &False, nil, nil, newVolumeError("Failed to create snapshot content with error failed to get input parameters to create snapshot snap7-3: \"failed to retrieve snapshot class  from the informer: \\\"volumesnapshotclass.snapshot.storage.k8s.io \\\\\\\"\\\\\\\" not found\\\"\""), false, true),
			initialClaims:         newClaimArray("claim7-3", "pvc-uid7-3", "1Gi", "volume7-3", v1.ClaimBound, &classEmpty),
			initialVolumes:        newVolumeArray("volume7-3", "pv-uid7-3", "pv-handle7-3", "1Gi", "pvc-uid7-3", "claim7-3", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
			initialStorageClasses: []*storage.StorageClass{diffDriverStorageClass},
			expectedEvents:        []string{"Warning SnapshotContentCreationFailed"},
			errors:                noerrors,
			expectSuccess:         false,
			test:                  testSyncSnapshot,
		},*/
		{
			name:              "7-4 - fail create snapshot with no-existing claim",
			initialContents:   nocontents,
			expectedContents:  nocontents,
			initialSnapshots:  newSnapshotArray("snap7-4", "snapuid7-4", "claim7-4", "", classGold, "", &False, nil, nil, nil, false, true),
			expectedSnapshots: newSnapshotArray("snap7-4", "snapuid7-4", "claim7-4", "", classGold, "", &False, nil, nil, newVolumeError("Failed to create snapshot content with error snapshot controller failed to update snap7-4 on API server: cannot get claim from snapshot"), false, true),
			initialVolumes:    newVolumeArray("volume7-4", "pv-uid7-4", "pv-handle7-4", "1Gi", "pvc-uid7-4", "claim7-4", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
			expectedEvents:    []string{"Warning SnapshotContentCreationFailed"},
			errors:            noerrors,
			expectSuccess:     false,
			test:              testSyncSnapshot,
		},
		{
			name:              "7-5 - fail create snapshot with no-existing volume",
			initialContents:   nocontents,
			expectedContents:  nocontents,
			initialSnapshots:  newSnapshotArray("snap7-5", "snapuid7-5", "claim7-5", "", classGold, "", &False, nil, nil, nil, false, true),
			expectedSnapshots: newSnapshotArray("snap7-5", "snapuid7-5", "claim7-5", "", classGold, "", &False, nil, nil, newVolumeError("Failed to create snapshot content with error failed to get input parameters to create snapshot snap7-5: \"failed to retrieve PV volume7-5 from the API server: \\\"cannot find volume volume7-5\\\"\""), false, true),
			initialClaims:     newClaimArray("claim7-5", "pvc-uid7-5", "1Gi", "volume7-5", v1.ClaimBound, &classEmpty),
			expectedEvents:    []string{"Warning SnapshotContentCreationFailed"},
			errors:            noerrors,
			expectSuccess:     false,
			test:              testSyncSnapshot,
		},

		{
			name:              "7-6 - fail create snapshot with claim that is not yet bound",
			initialContents:   nocontents,
			expectedContents:  nocontents,
			initialSnapshots:  newSnapshotArray("snap7-6", "snapuid7-6", "claim7-6", "", classGold, "", &False, nil, nil, nil, false, true),
			expectedSnapshots: newSnapshotArray("snap7-6", "snapuid7-6", "claim7-6", "", classGold, "", &False, nil, nil, newVolumeError("Failed to create snapshot content with error failed to get input parameters to create snapshot snap7-6: \"the PVC claim7-6 is not yet bound to a PV, will not attempt to take a snapshot\""), false, true),
			initialClaims:     newClaimArray("claim7-6", "pvc-uid7-6", "1Gi", "", v1.ClaimPending, &classEmpty),
			expectedEvents:    []string{"Warning SnapshotContentCreationFailed"},
			errors:            noerrors,
			expectSuccess:     false,
			test:              testSyncSnapshot,
		},
		/*{
				name:              "7-8 - fail create snapshot due to cannot update snapshot status",
				initialContents:   nocontents,
				expectedContents:  nocontents,
				initialSnapshots:  newSnapshotArray("snap7-8", "snapuid7-8", "claim7-8", "", classGold, "", &False, nil, nil, nil),
				expectedSnapshots: newSnapshotArray("snap7-8", "snapuid7-8", "claim7-8", "", classGold, "", &False, nil, nil, newVolumeError("Failed to create snapshot: snapshot controller failed to update default/snap7-8 on API server: mock update error")),
				initialClaims:     newClaimArray("claim7-8", "pvc-uid7-8", "1Gi", "volume7-8", v1.ClaimBound, &classEmpty),
				initialVolumes:    newVolumeArray("volume7-8", "pv-uid7-8", "pv-handle7-8", "1Gi", "pvc-uid7-8", "claim7-8", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
				expectedCreateCalls: []createCall{
								{
								snapshotName: "snapshot-snapuid7-8",
								volume:       newVolume("volume7-8", "pv-uid7-8", "pv-handle7-8", "1Gi", "pvc-uid7-8", "claim7-8", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
								parameters:   map[string]string{"param1": "value1"},
								// information to return
								driverName:   mockDriverName,
								size:         defaultSize,
								snapshotId:   "sid7-8",
								creationTime: timeNow,
								readyToUse:   True,
						},
				},
				/*errors: []reactorError{
								// Inject error to the forth client.VolumesnapshotV1beta1().VolumeSnapshots().Update call.
								// All other calls will succeed.
								{"update", "volumesnapshots", errors.New("mock update error")},
								{"update", "volumesnapshots", errors.New("mock update error")},
								{"update", "volumesnapshots", errors.New("mock update error")},
						},
						expectedEvents: []string{"Warning SnapshotContentCreationFailed"},

						expectSuccess:     false,
						test:           testSyncSnapshot,
				},
		        /*{
		                // TODO(xiangqian): this test case needs to be revisited the scenario
		                // of VolumeSnapshotContent saving failure. Since there will be no content object
		                // in API server, it could potentially cause leaking issue
		                name:              "7-9 - fail create snapshot due to cannot save snapshot content",
		                initialContents:   nocontents,
		                expectedContents:  nocontents,
		                initialSnapshots:  newSnapshotArray("snap7-9", "snapuid7-9", "claim7-9", "", classGold, "", &False, nil, nil, nil),
		                expectedSnapshots: newSnapshotArray("snap7-9", "snapuid7-9", "claim7-9", "", classGold, "snapcontent-snapuid7-9", &True, metaTimeNowUnix, getSize(defaultSize), nil),
		                initialClaims:     newClaimArray("claim7-9", "pvc-uid7-9", "1Gi", "volume7-9", v1.ClaimBound, &classEmpty),
		                initialVolumes:    newVolumeArray("volume7-9", "pv-uid7-9", "pv-handle7-9", "1Gi", "pvc-uid7-9", "claim7-9", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
		                expectedCreateCalls: []createCall{
		                        {
		                                snapshotName: "snapshot-snapuid7-9",
		                                volume:       newVolume("volume7-9", "pv-uid7-9", "pv-handle7-9", "1Gi", "pvc-uid7-9", "claim7-9", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
		                                parameters:   map[string]string{"param1": "value1"},
		                                // information to return
		                                driverName:   mockDriverName,
		                                size:         defaultSize,
		                                snapshotId:   "sid7-9",
		                                creationTime: timeNow,
		                                readyToUse:   True,
		                        },
		                },
		                errors: []reactorError{
		                        {"create", "volumesnapshotcontents", errors.New("mock create error")},
		                        {"create", "volumesnapshotcontents", errors.New("mock create error")},
		                        {"create", "volumesnapshotcontents", errors.New("mock create error")},
		                },
		                expectedEvents: []string{"Warning CreateSnapshotContentFailed"},
		                test:           testSyncSnapshot,
		        },
				{
					name:              "7-10 - fail create snapshot with secret not found",
					initialContents:   nocontents,
					expectedContents:  nocontents,
					initialSnapshots:  newSnapshotArray("snap7-10", "snapuid7-10", "claim7-10", "", validSecretClass, "", &False, nil, nil, nil, false, true),
					expectedSnapshots: newSnapshotArray("snap7-10", "snapuid7-10", "claim7-10", "", validSecretClass, "", &False, nil, nil, newVolumeError("Failed to create snapshot: error getting secret secret in namespace default: cannot find secret secret"), false, true),
					initialClaims:     newClaimArray("claim7-10", "pvc-uid7-10", "1Gi", "volume7-10", v1.ClaimBound, &classEmpty),
					initialVolumes:    newVolumeArray("volume7-10", "pv-uid7-10", "pv-handle7-10", "1Gi", "pvc-uid7-10", "claim7-10", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, classEmpty),
					initialSecrets:    []*v1.Secret{}, // no initial secret created
					errors:            noerrors,
					test:              testSyncSnapshot,
				},*/
	}
	runSyncTests(t, tests, snapshotClasses)
}