# Recipes

Step-by-step guides for common changes. Each one ends with `make check`.

## Add an HTTP endpoint

1. Put the logic in a business package (see [below](#add-a-business-package)),
   or add a method to an existing one. Return sentinel errors for expected failures.
2. If the handler needs a new method, extend the interface declared in
   `internal/httpapi/handler.go` (or declare a new one there).
3. Write the handler in `internal/httpapi` following `greet`:

   ```go
   type createOrderRequest struct {
   	Item string `json:"item"`
   }

   func (h *handlers) createOrder(w http.ResponseWriter, r *http.Request) {
   	req, err := decode[createOrderRequest](w, r)
   	if err != nil {
   		h.respondError(w, r, http.StatusBadRequest, err)
   		return
   	}

   	order, err := h.orders.Create(r.Context(), req.Item)
   	if err != nil {
   		h.respondError(w, r, statusFor(err), err)
   		return
   	}

   	h.respond(w, r, http.StatusCreated, order)
   }
   ```

4. Register the route in `NewHandler`: `mux.HandleFunc("POST /v1/orders", h.createOrder)`.
   Path parameters: `"GET /v1/orders/{id}"` and `r.PathValue("id")`.
5. Map new business errors in `statusFor`.
6. Add cases to `internal/httpapi/handler_test.go`.

## Add a gRPC method

1. Edit `api/greeter/v1/greeter.proto` (or add `api/<name>/v1/<name>.proto`).
   Follow buf `STANDARD` lint rules: `FooRequest`/`FooResponse` messages,
   `FooService` service names, versioned packages.
2. Run `make generate` and `make breaking`.
3. Implement the method in `internal/grpcapi` (for a new service: a new struct
   embedding `Unimplemented...Server`, registered in `NewServer`).
   Use getters (`req.GetName()`) and builders (`pb.FooResponse_builder{...}.Build()`).
4. Map new business errors in `toStatus`.
5. Add cases to `internal/grpcapi/server_test.go`.

## Add a config option

1. Add a field to the right struct in `internal/config/config.go` with
   `default`, `env` and `help` tags. Group related options in a nested struct
   with `embed:"" prefix:"x-" envprefix:"X_"`.
2. Pass the value from `internal/app` to the code that needs it
   (as a constructor argument, not the whole `Config`).
3. Add it to `.env.example` and, if it differs in production, to `chart/values.yaml` (`env:`).
4. Cover parsing in `internal/config/config_test.go` if it has non-trivial validation.

## Add a business package

1. Create `internal/<name>/` with a package comment explaining its role.
2. Export a `Service` (or a more specific name) with a `New(...)` constructor
   taking its dependencies explicitly.
3. Export sentinel errors for expected failures.
4. No imports of `httpapi`, `grpcapi`, `config` or `app`.
5. Construct it in `app.New` and pass it to the transports through their `Deps`.
6. Unit test it without any transport.

## Add a dependency (database, cache, client, ...)

Example: PostgreSQL.

1. Add connection options to `internal/config` (e.g. `DB.URL` with `env:"DB_URL"`).
2. Create the client in `app.New`; return an error if it fails, so the
   service does not start half-configured.
3. Wrap data access in its own package (e.g. `internal/storage`), exposing
   methods that business packages call through interfaces they declare.
4. Close the client in `App.shutdown`, after the servers have stopped.
5. If the dependency is required to serve traffic, include its health in
   readiness (`ready` function passed to `debug.NewHandler`).
6. Register its metrics in the Prometheus registry created in `app.New`.
7. Add the service to `compose.yaml` for local development and document the
   new variables in `.env.example` and `chart/values.yaml`.
