.PHONY: validate gen gen-go gen-ts clean

validate:
	pnpm exec redocly lint openapi.yaml

gen-go:
	@mkdir -p ../../apps/backend/internal/api
	cd ../../apps/backend && go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 \
		-package api \
		-generate types,chi-server,strict-server,spec \
		-o internal/api/api.gen.go \
		../../packages/api-schema/openapi.yaml

gen-ts:
	@mkdir -p ../../apps/frontend/src/lib/api
	pnpm exec openapi-typescript openapi.yaml \
		-o ../../apps/frontend/src/lib/api/types.gen.ts

gen: validate gen-go gen-ts

clean:
	rm -f ../../apps/backend/internal/api/api.gen.go
	rm -f ../../apps/frontend/src/lib/api/types.gen.ts
