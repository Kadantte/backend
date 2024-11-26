# Copyright (C) 2024 Pavel Sobolev
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published
# by the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

FROM docker.io/golang:1.23.2-alpine AS builder

WORKDIR /build/src

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    --mount=type=bind,target=. \
    go build -v -ldflags "-s -w" -trimpath -o /build/morbo

FROM alpine:3.20

LABEL org.opencontainers.image.source=https://github.com/morbo-org/Morbo

EXPOSE 80

ARG USER=morbo
ENV HOME=/home/$USER

RUN adduser -D $USER

COPY --from=builder /build/morbo /usr/local/bin

USER $USER
WORKDIR $HOME

RUN mkdir -p ~/.local/share/morbo

ENTRYPOINT [ "morbo" ]
