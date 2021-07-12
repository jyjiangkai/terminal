#
# Copyright 2017 Huawei Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#
# This file is used for building and testing agent.
# It includes some useful tools for development of agent
# which can ensure everyone using similar development toolkit

FROM hub.easystack.io/production/golang:1.15.6

ENV NVIDIA_VISIBLE_DEVICES=all \
    NVIDIA_DRIVER_CAPABILITIES=utility \
    TERMINAL=terminal

ARG binary=${GOPATH}/src/${TERMINAL}/bin/${TERMINAL}

WORKDIR ${GOPATH}/src/${TERMINAL}

RUN cp ${binary} /usr/bin/${TERMINAL}

ENTRYPOINT ["/usr/bin/${TERMINAL}"]
