// Copyright IBM Corp. 2023, 2025
// SPDX-License-Identifier: MPL-2.0

package rate

import (
	"bytes"
	"sync"
)

var keyBuilderPool sync.Pool

type builder struct {
	Buffer bytes.Buffer
}

func join(parts ...string) string {
	var b *builder
	if v := keyBuilderPool.Get(); v != nil {
		b = v.(*builder)
		b.Buffer.Reset()
	} else {
		b = &builder{}
	}
	defer keyBuilderPool.Put(b)

	end := len(parts) - 1
	for i, p := range parts {
		b.Buffer.WriteString(p)
		if i != end {
			b.Buffer.Write([]byte(":"))
		}
	}
	return b.Buffer.String()
}
