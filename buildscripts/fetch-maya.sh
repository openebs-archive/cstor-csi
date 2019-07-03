CV_PATH=pkg/cstor/volume/v1alpha1
MAPIS_PATH=pkg/apis/openebs.io/maya/v1alpha1
GITHUB_MAYA=https://raw.githubusercontent.com/openebs/maya/master

wget -P $MAPIS_PATH $GITHUB_MAYA/pkg/apis/openebs.io/v1alpha1/cstor_volume.go 
wget -P $CV_PATH  $GITHUB_MAYA/$CV_PATH/build.go
wget -P $CV_PATH  $GITHUB_MAYA/$CV_PATH/kubernetes.go
wget -P $CV_PATH  $GITHUB_MAYA/$CV_PATH/cstorvolume.go
