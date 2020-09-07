# Copyright © 2020 The OpenEBS Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#!/bin/bash

# Used by development setups especially Vagrant
# NOTE: Use of vagrant user !!!
set -ex

GO_VERSION="1.9.1"
CURDIR=`pwd`

# Setup go, for development
SRCROOT="/opt/go"
SRCPATH="/opt/gopath"

# Get the ARCH
ARCH=`uname -m | sed 's|i686|386|' | sed 's|x86_64|amd64|'`

# Install Go
cd /tmp

if [ ! -f "./go${GO_VERSION}.linux-${ARCH}.tar.gz" ]; then
  wget -q https://storage.googleapis.com/golang/go${GO_VERSION}.linux-${ARCH}.tar.gz
fi

tar -xf go${GO_VERSION}.linux-${ARCH}.tar.gz
sudo mv go $SRCROOT
sudo chmod 775 $SRCROOT
sudo chown vagrant:vagrant $SRCROOT

# Setup the GOPATH; 
# This allows subsequent "go get" commands to work.
sudo mkdir -p $SRCPATH
sudo chown -R vagrant:vagrant $SRCPATH 2>/dev/null || true
# ^^ silencing errors here because we expect this to fail for the shared folder

cat <<EOF >/tmp/gopath.sh
export GOPATH="$SRCPATH"
export GOROOT="$SRCROOT"
export PATH="$SRCROOT/bin:$SRCPATH/bin:\$PATH"
EOF

sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh
sudo chmod 0755 /etc/profile.d/gopath.sh
source /etc/profile.d/gopath.sh

cd ${CURDIR}
