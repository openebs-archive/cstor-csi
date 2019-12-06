package iscsi

import "testing"

func TestDiscovery(t *testing.T) {
	execCommand = fakeExecCommand
	tests := map[string]struct {
		tgtPortal        string
		iface            string
		discoverySecret  Secrets
		chapDiscovery    bool
		wantErr          bool
		mockedStdout     string
		mockedExitStatus int
	}{
		"DiscoverySuccess": {
			tgtPortal:        "172.18.0.2:3260",
			iface:            "default",
			chapDiscovery:    false,
			mockedStdout:     "172.18.0.2:3260,1 iqn.2016-09.com.openebs.jiva:store1\n",
			mockedExitStatus: 0,
		},

		"ConnectionFailure": {
			tgtPortal:     "172.18.0.2:3262",
			iface:         "default",
			chapDiscovery: false,
			mockedStdout: `iscsiadm: cannot make connection to 172.18.0.2: Connection refused
iscsiadm: cannot make connection to 172.18.0.2: Connection refused
iscsiadm: connection login retries (reopen_max) 5 exceeded
iscsiadm: Could not perform SendTargets discovery: encountered connection failure\n`,
			mockedExitStatus: 4,
			wantErr:          true,
		},

		"ChapEntrySuccess": {
			tgtPortal:     "172.18.0.2:3260",
			iface:         "default",
			chapDiscovery: true,
			discoverySecret: Secrets{
				UserNameIn: "dummyuser",
				PasswordIn: "dummypass",
			},
			mockedStdout:     "172.18.0.2:3260,1 iqn.2016-09.com.openebs.jiva:store1\n",
			mockedExitStatus: 0,
		},

		"ChapEntryFailure": {
			tgtPortal: "172.18.0.2:3260",
			iface:     "default",
			discoverySecret: Secrets{
				UserNameIn: "dummyuser",
				PasswordIn: "dummypass",
			},
			chapDiscovery: true,
			mockedStdout: `iscsiadm: Login failed to authenticate with target
iscsiadm: discovery login to 172.18.0.2 rejected: initiator error (02/01), non-retryable, giving up
iscsiadm: Could not perform SendTargets discovery.\n`,
			mockedExitStatus: 4,
			wantErr:          true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockedExitStatus = tt.mockedExitStatus
			mockedStdout = tt.mockedStdout
			err := Discovery(tt.tgtPortal, tt.iface, tt.discoverySecret, tt.chapDiscovery)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discovery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCreateDBEntry(t *testing.T) {
	execCommand = fakeExecCommand
	tests := map[string]struct {
		tgtPortal        string
		tgtIQN           string
		iface            string
		discoverySecret  Secrets
		sessionSecret    Secrets
		chapDiscovery    bool
		wantErr          bool
		mockedStdout     string
		mockedExitStatus int
	}{
		"CreateDBEntrySuccess": {
			tgtPortal:        "192.168.1.107:3260",
			tgtIQN:           "iqn.2010-10.org.openstack:volume-eb393993-73d0-4e39-9ef4-b5841e244ced",
			iface:            "default",
			chapDiscovery:    false,
			mockedStdout:     nodeDB,
			mockedExitStatus: 0,
		},
		"CreateDBEntryWithChapDiscoverySuccess": {
			tgtPortal: "192.168.1.107:3260",
			tgtIQN:    "iqn.2010-10.org.openstack:volume-eb393993-73d0-4e39-9ef4-b5841e244ced",
			iface:     "default",
			discoverySecret: Secrets{
				UserNameIn:  "dummyuser",
				PasswordIn:  "dummypass",
				SecretsType: "chap",
			},
			sessionSecret: Secrets{
				UserNameIn:  "dummyuser",
				PasswordIn:  "dummypass",
				SecretsType: "chap",
			},
			chapDiscovery:    true,
			mockedStdout:     nodeDB,
			mockedExitStatus: 0,
		},
		"CreateDBEntryWithChapDiscoveryFailure": {
			tgtPortal:        "172.18.0.2:3260",
			tgtIQN:           "iqn.2016-09.com.openebs.jiva:store1",
			iface:            "default",
			chapDiscovery:    true,
			mockedStdout:     "iscsiadm: No records found\n",
			mockedExitStatus: 21,
			wantErr:          true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockedExitStatus = tt.mockedExitStatus
			mockedStdout = tt.mockedStdout
			err := CreateDBEntry(tt.tgtIQN, tt.tgtPortal, tt.iface, tt.discoverySecret, tt.sessionSecret, tt.chapDiscovery)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDBEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

}
