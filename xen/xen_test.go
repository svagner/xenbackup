package xen

import (
	"flag"
	"strings"
	"testing"

	xapi "github.com/svagner/go-xenserver-client"
)

var host = flag.String("host", "", "Hypervisor's name")
var username = flag.String("user", "", "Hypervisor's username")
var password = flag.String("pass", "", "Hypervisor's password")
var vmname = flag.String("vmname", "", "Virtual machine's name")
var vmuuid = flag.String("vmuuid", "", "Virtual machine's uuid")
var path = flag.String("path", "", "Path for save backups")

func TestCreateDeleteSnapshot(t *testing.T) {
	flag.Parse()
	var backupXen XenBackup
	err := backupXen.Init(*host, *username, *password)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	backupXen.VMs = append(backupXen.VMs, VM{Name: *vmname, Uuid: *vmuuid})
	vms := make([]*xapi.VM, 0)
	for _, vm := range backupXen.VMs {
		if vm.Uuid != "" {
			node, err := backupXen.client.GetVMByUuid(vm.Uuid)
			if err != nil {
				t.Log(err)
				t.FailNow()
			}
			vms = append(vms, node)
			continue
		}
		if vm.Name != "" {
			res, err := backupXen.client.GetVMByNameLabel(vm.Name)
			if err != nil {
				t.Log(err)
				t.FailNow()
			}
			if len(res) == 0 {
				t.Log("VM not found by name ", vm.Name)
				t.FailNow()
			}
			for _, node := range res {
				vms = append(vms, node)
			}
		}
	}
	// Start doing snapshots
	for _, vm := range vms {
		name, err := vm.GetName()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		backupName := backupXen.generateBackupName(name)
		t.Log("Start to create snapshot", backupName)
		snapshot, err := vm.Snapshot(backupName)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		err = snapshot.SetTemplateAsVM(false)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		uuid, err := snapshot.GetUuid()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		t.Log("UUID snapshot:", uuid)
		snapshotName, err := snapshot.GetName()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		if strings.HasPrefix(snapshotName, SNAPSHOT_PREFIX) {
			t.Log("Trying to delete snapshot ", snapshotName)
			t.Log("Get snapshot's list:")
			disks, err := snapshot.GetDisks()
			if err != nil {
				t.Log(err)
				t.FailNow()
			}
			t.Log("Snapshot", snapshotName, "disks:")
			for idx, disk := range disks {
				u, e := disk.GetUuid()
				if e != nil {
					t.Log(e)
					t.FailNow()
				}
				t.Log("UUID", idx, u)
			}
			// first step - delete snapshot
			err = snapshot.Destroy()
			if err != nil {
				t.Log("Error while destroy snapshot " + backupName + ": " + err.Error())
				t.FailNow()
			}
			// as second step - we have to delete vdi drive os snapshot
			t.Log("Delete vdi's for snapshot " + backupName)
			for idx, disk := range disks {
				u, e := disk.GetUuid()
				if e != nil {
					t.Log(e)
					t.FailNow()
				}
				t.Log("Deleting vdi ", idx, u)
				e = disk.Destroy()
				if e != nil {
					t.Log("Delete vdi", u, "failed:", e)
					t.FailNow()
				}
			}

		} else {
			t.Log("VM image wasn't delete: name '" + name + "' hasn't prefix '" + SNAPSHOT_PREFIX + "'")
			t.FailNow()
		}
	}
}

func TestCreateSnapshot(t *testing.T) {
	flag.Parse()
	var backupXen XenBackup
	err := backupXen.Init(*host, *username, *password)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	backupXen.VMs = append(backupXen.VMs, VM{Name: *vmname, Uuid: *vmuuid})
	vms := make([]*xapi.VM, 0)
	for _, vm := range backupXen.VMs {
		if vm.Uuid != "" {
			node, err := backupXen.client.GetVMByUuid(vm.Uuid)
			if err != nil {
				t.Log(err)
				t.FailNow()
			}
			vms = append(vms, node)
			continue
		}
		if vm.Name != "" {
			res, err := backupXen.client.GetVMByNameLabel(vm.Name)
			if err != nil {
				t.Log(err)
				t.FailNow()
			}
			if len(res) == 0 {
				t.Log("VM not found by name ", vm.Name)
				t.FailNow()
			}
			for _, node := range res {
				vms = append(vms, node)
			}
		}
	}
	// Start doing snapshots
	for _, vm := range vms {
		name, err := vm.GetName()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		backupName := backupXen.generateBackupName(name)
		t.Log("Start to create snapshot", backupName)
		snapshot, err := vm.Snapshot(backupName)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		err = snapshot.SetTemplateAsVM(false)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		uuid, err := snapshot.GetUuid()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		t.Log("UUID snapshot:", uuid)
		snapshotName, err := snapshot.GetName()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		t.Log("Snapshot name:", snapshotName)
		disks, err := snapshot.GetDisks()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
		t.Log("Snapshot", snapshotName, "disks:")
		for idx, disk := range disks {
			u, e := disk.GetUuid()
			if e != nil {
				t.Log(e)
				t.FailNow()
			}
			t.Log("UUID", idx, u)
		}
	}
}
