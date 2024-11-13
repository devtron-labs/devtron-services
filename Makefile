
TARGET_BRANCH?=7bd23ef50c96c1677aeb13fb226d534ac636e1bc

all: build

dep-update-oss:
	cd git-sensor && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd chart-sync && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd ci-runner && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd kubelink && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd kubewatch && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss
	cd lens && TARGET_BRANCH=$(TARGET_BRANCH) $(MAKE) dep-update-oss

build:
	cd chart-sync && $(MAKE)
	cd ci-runner && $(MAKE)
#	cd devtctl && $(MAKE)
	cd git-sensor && $(MAKE)
	cd kubelink && $(MAKE)
	cd kubewatch && $(MAKE)
	cd lens && $(MAKE)
#	cd common-lib && $(MAKE)