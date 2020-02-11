# Changelog

## [1.7.0](https://github.com/openebs/cstor-csi/tree/1.7.0) (2020-02-15)

**Merged pull requests:**

- fix\(cstor-csi\): dealock in monitor mount goroutine [\#72](https://github.com/openebs/cstor-csi/pull/72) ([@utkarshmani1997](https://github.com/utkarshmani1997))
- fix\(csi, snapshot\): enable status subresource in VolumeSnapshot CRD [\#71](https://github.com/openebs/cstor-csi/pull/71) ([@prateekpandey14](https://github.com/prateekpandey14))
- fix\(csi\): handle duplicate snapshot request and support Snapshot V1beta1 APIs [\#70](https://github.com/openebs/cstor-csi/pull/70) ([@prateekpandey14](https://github.com/prateekpandey14))
- feat\(topology\): Add topology support for cstor csi volumes [\#69](https://github.com/openebs/cstor-csi/pull/69) ([@payes](https://github.com/payes))

- fix(BDD): wait for restarted pod to come to running state ([#1608](https://github.com/openebs/maya/pull/1608),
  [@shubham14bajpai](https://github.com/shubham14bajpai))


## [1.6.0](https://github.com/openebs/cstor-csi/tree/1.6.0) (2020-01-15)

**Merged pull requests:**

- refact\(version\): bump branch version to 1.6.0 [\#68](https://github.com/openebs/cstor-csi/pull/68) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(version\): bump version to 1.7.0 [\#67](https://github.com/openebs/cstor-csi/pull/67) ([@prateekpandey14](https://github.com/prateekpandey14))
- feat\(csi\): get cstorvolume policy from the SC parameters [\#66](https://github.com/openebs/cstor-csi/pull/66) ([@prateekpandey14](https://github.com/prateekpandey14))
- feat\(csi\): add block volume support for cstor volume [\#65](https://github.com/openebs/cstor-csi/pull/65) ([@prateekpandey14](https://github.com/prateekpandey14))
- feat\(csi-metrics\): add metrics support for volumes [\#64](https://github.com/openebs/cstor-csi/pull/64) ([@prateekpandey14](https://github.com/prateekpandey14))
- fix\(unmount\): Fix unmount errors for xfs filesystem [\#63](https://github.com/openebs/cstor-csi/pull/63) ([@payes](https://github.com/payes))
- fix\(resize\): Add command to resize xfs volumes [\#61](https://github.com/openebs/cstor-csi/pull/61) ([@payes](https://github.com/payes))
- fix\(remount\): Avoid monitoring volume unless mounted [\#60](https://github.com/openebs/cstor-csi/pull/60) ([@payes](https://github.com/payes))

## [1.5.0](https://github.com/openebs/cstor-csi/tree/1.5.0) (2019-12-15)

**Merged pull requests:**

- cherry-pick to 1.5.x [\#62](https://github.com/openebs/cstor-csi/pull/62) ([@payes](https://github.com/payes))
- refact\(size\): conversion G to Gi [\#59](https://github.com/openebs/cstor-csi/pull/59) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(cstor-csi\): Modularise CSI implementation  [\#58](https://github.com/openebs/cstor-csi/pull/58) ([@payes](https://github.com/payes))
- Add support for xfs filesystem [\#57](https://github.com/openebs/cstor-csi/pull/57) ([@payes](https://github.com/payes))
- Add\(env\): Add environment variable to enable remount feature [\#56](https://github.com/openebs/cstor-csi/pull/56) ([@payes](https://github.com/payes))
- Update\(Readme\): Update steps to clone cstor volumes [\#55](https://github.com/openebs/cstor-csi/pull/55) ([@payes](https://github.com/payes))
- refact\(docs\): update storage requests unit size [\#54](https://github.com/openebs/cstor-csi/pull/54) ([@ranjithwingrider](https://github.com/ranjithwingrider))

## [1.4.0](https://github.com/openebs/cstor-csi/tree/1.4.0) (2019-11-15)

**Merged pull requests:**

- fix\(csi-node\): return error for orphaned volume during unpublish event [\#53](https://github.com/openebs/cstor-csi/pull/53) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(driver\): update driver name to cstor.csi.openebs.io [\#52](https://github.com/openebs/cstor-csi/pull/52) ([@prateekpandey14](https://github.com/prateekpandey14))
- fix\(csi-node\): return error for orphaned volume during unpublish event [\#51](https://github.com/openebs/cstor-csi/pull/51) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(driver\): update driver name to cstor.csi.openebs.io [\#50](https://github.com/openebs/cstor-csi/pull/50) ([@prateekpandey14](https://github.com/prateekpandey14))
- fix\(Makefile\): update image name in push target [\#47](https://github.com/openebs/cstor-csi/pull/47) ([@prateekpandey14](https://github.com/prateekpandey14))
- feat\(snapshot\): Implement APIs to Create/Delete Snapshots [\#46](https://github.com/openebs/cstor-csi/pull/46) ([@payes](https://github.com/payes))
- refactor\(code\): update import paths from openebs/csi to openebs/cstor-csi [\#45](https://github.com/openebs/cstor-csi/pull/45) ([@payes](https://github.com/payes))

## [1.3.0](https://github.com/openebs/cstor-csi/tree/1.3.0) (2019-10-15)

**Merged pull requests:**

- refact\(travis\): remove travis github release deploy [\#49](https://github.com/openebs/cstor-csi/pull/49) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(travis\): remove travis github release deploy [\#48](https://github.com/openebs/cstor-csi/pull/48) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(travis\): remove travis github release deploy [\#44](https://github.com/openebs/cstor-csi/pull/44) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(cvc\): populate version details for csi based volumes \(\#41\) [\#42](https://github.com/openebs/cstor-csi/pull/42) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(cvc\): populate version details for csi based volumes [\#41](https://github.com/openebs/cstor-csi/pull/41) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(travis\): bump minikube version v1.4.0 to use k8s v1.16.0 [\#40](https://github.com/openebs/cstor-csi/pull/40) ([@prateekpandey14](https://github.com/prateekpandey14))
- refact\(travis\): bump minikube version 1.4.0 to use k8s v1.16.0 [\#39](https://github.com/openebs/cstor-csi/pull/39) ([@prateekpandey14](https://github.com/prateekpandey14))
- update\(yamls\): Update example deployment yamls [\#34](https://github.com/openebs/cstor-csi/pull/34) ([@payes](https://github.com/payes))

## [1.2.0](https://github.com/openebs/cstor-csi/tree/1.2.0) (2019-08-15)

**Merged pull requests:**

- update\(readme\): Update readme based on the recent csi changes [\#36](https://github.com/openebs/cstor-csi/pull/36) ([@payes](https://github.com/payes))
- fix\(travis\): update travis api secure key [\#35](https://github.com/openebs/cstor-csi/pull/35) ([@prateekpandey14](https://github.com/prateekpandey14))
- refactor\(publish/unpublish\): Refactor code for node publish and unpublish [\#32](https://github.com/openebs/cstor-csi/pull/32) ([@payes](https://github.com/payes))
- refact\(cvc\): use cstorPoolCluster name instead of storagePoolClaim [\#30](https://github.com/openebs/cstor-csi/pull/30) ([@prateekpandey14](https://github.com/prateekpandey14))
- feat\(resize\): Implement Volume Expansion Endpoints [\#13](https://github.com/openebs/cstor-csi/pull/13) ([@payes](https://github.com/payes))

## [1.1.0](https://github.com/openebs/cstor-csi/tree/1.1.0) (2019-07-15)

**Fixed bugs:**

- fix\(version\): update Version file path and correct a typo [\#6](https://github.com/openebs/cstor-csi/pull/6) ([@payes](https://github.com/payes))
- fix\(build\): updates buildscript to generate driver image [\#5](https://github.com/openebs/cstor-csi/pull/5) ([@AmitKumarDas](https://github.com/AmitKumarDas))

**Merged pull requests:**

- docs\(readme\): update usage instructions for OpenEBS 1.1.0 [\#31](https://github.com/openebs/cstor-csi/pull/31) ([@kmova](https://github.com/kmova))
- Update\(README\): Add iscsi client package as a prerequisite [\#29](https://github.com/openebs/cstor-csi/pull/29) ([@payes](https://github.com/payes))
- Merge changes to 1.1 branch [\#28](https://github.com/openebs/cstor-csi/pull/28) ([@payes](https://github.com/payes))
- Add Status field to CSIVolume API and read CSIVolumeCRs while coming up [\#27](https://github.com/openebs/cstor-csi/pull/27) ([@payes](https://github.com/payes))
- Update\(operator\): Add iscsiadm file and related library mounts to operator yaml [\#26](https://github.com/openebs/cstor-csi/pull/26) ([@payes](https://github.com/payes))
- refact\(cvc\):remove configclass reference from CVC and add validations [\#25](https://github.com/openebs/cstor-csi/pull/25) ([@prateekpandey14](https://github.com/prateekpandey14))
- fix\(golang-ci\): update folder structure and fix golang-ci lint error [\#24](https://github.com/openebs/cstor-csi/pull/24) ([@prateekpandey14](https://github.com/prateekpandey14))
- chore\(operator\): add missing resources and update required clusterrole [\#23](https://github.com/openebs/cstor-csi/pull/23) ([@prateekpandey14](https://github.com/prateekpandey14))
- chore\(cvc\): refactor cstorvolumeclaim schema [\#21](https://github.com/openebs/cstor-csi/pull/21) ([@prateekpandey14](https://github.com/prateekpandey14))
- update\(README\): Add steps to provision a volume [\#20](https://github.com/openebs/cstor-csi/pull/20) ([@payes](https://github.com/payes))
- chore\(config\): add sample yamls for creating CSI based volumes [\#19](https://github.com/openebs/cstor-csi/pull/19) ([@payes](https://github.com/payes))
- feat\(ProvisionVolume\): Update Volume Provisioning and Deletion using CVC [\#18](https://github.com/openebs/cstor-csi/pull/18) ([@payes](https://github.com/payes))
- Add cstor volume builder pattern [\#17](https://github.com/openebs/cstor-csi/pull/17) ([@payes](https://github.com/payes))
- Add additional builder patterns for csi volume [\#15](https://github.com/openebs/cstor-csi/pull/15) ([@payes](https://github.com/payes))
- feat\(cvc\): Add CStorVolumeClaim and, corresponding clientset and builder pattern [\#14](https://github.com/openebs/cstor-csi/pull/14) ([payes](https://github.com/payes))
- chore\(idiomatic\): rename functions, structures to follow idiomatic style [\#12](https://github.com/openebs/cstor-csi/pull/12) ([@AmitKumarDas](https://github.com/AmitKumarDas))
- Run CSI BDD tests in travis [\#11](https://github.com/openebs/cstor-csi/pull/11) ([@payes](https://github.com/payes))
- feat\(csi, operator\): add csi driver, attacher, node-plugin, and conroller [\#10](https://github.com/openebs/cstor-csi/pull/10) ([@prateekpandey14](https://github.com/prateekpandey14))
- refactor\(idiomatic\): update provisioning code to follow standard practices [\#9](https://github.com/openebs/cstor-csi/pull/9) ([@AmitKumarDas](https://github.com/AmitKumarDas))
- refactor\(driver\): update driver code to follow idiomatic style [\#8](https://github.com/openebs/cstor-csi/pull/8) ([@AmitKumarDas](https://github.com/AmitKumarDas))
- fix\(travis\): update travis build [\#7](https://github.com/openebs/cstor-csi/pull/7) ([@prateekpandey14](https://github.com/prateekpandey14))
- chore\(compile, vendor\): add required vendoring packages [\#4](https://github.com/openebs/cstor-csi/pull/4) ([@AmitKumarDas](https://github.com/AmitKumarDas))
- chore\(lint\): add travis, bettercodehub and golangci integrations [\#3](https://github.com/openebs/cstor-csi/pull/3) ([@payes](https://github.com/payes))
- Add cmd and pkg directories [\#2](https://github.com/openebs/cstor-csi/pull/2) ([@payes](https://github.com/payes))
- Add design and usage docs [\#1](https://github.com/openebs/cstor-csi/pull/1) ([@payes](https://github.com/payes))
