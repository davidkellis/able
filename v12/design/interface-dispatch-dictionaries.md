# Interpreter and Dynamic-Boundary Interface Dictionaries (v12)

## Status

This is the active execution record for interface values used by the Go
tree-walker, bytecode VM, and explicit runtime-value boundaries. The semantic
contract is spec §10.3.4: an upcast captures the implementation selected by
the impls visible at that cast site, including inherited interface members and
default methods.

It is deliberately not the ABI for a statically resolved compiled call. Those
calls use the direct Go interface carriers and adapters described by
`compiler-go-lowering-spec.md` and
`compiler-native-lowering-guardrails.md`.

## Interpreter value model

`runtime.InterfaceValue` holds the interface definition, the underlying
value, fully bound interface arguments, and its dispatch state:

- `SharedMethods` is a reusable raw-method dictionary;
- `Methods` is an optional per-value overlay, so a value never mutates a
  shared dictionary; and
- `BoundMethodName`/`BoundMethod` memoize the most recently bound receiver
  method for that particular interface value.

The dictionary contains the applicable implementation methods and interface
defaults, including methods from composite/base interfaces. Normal member
access looks in the value overlay and shared dictionary, binds the selected
method to the underlying receiver when needed, then stores only that bound
result in the per-value memo. This preserves the cast-site interface view
while avoiding a dictionary clone or repeated receiver binding for ordinary
repeated calls.

Generic interface arguments are part of both coercion and lookup. The
typechecker records inferred call arguments in the AST; the interpreter then
uses the fully bound interface/call information for method resolution and
return/constraint handling. An interface name with unbound parameters is not a
runtime interface value (spec §4.1 and §10).

Implementation selection happens once, at the upcast. A statically known
source with incomparable unnamed implementations is rejected before dispatch;
an unknown source checked by `as Interface` applies the same rule at runtime.
Generic method arguments instantiate the captured slot rather than selecting
another implementation. Default slots likewise retain the explicit override
or default selected for that interface view.

An interface method declared to return `Self` returns the same fully bound
interface view. The tree-walker and bytecode VM attach the originating
interface definition, arguments, shared dictionary, and an isolated copy of
any per-value overlay to the concrete result. This prevents a consumer package
from resolving the result against a different visible implementation.

## Construction and invalidation

For normal interpreter coercion, `coerceToInterfaceValue` verifies that the
concrete value implements the requested interface, then builds the method
dictionary from the selected impl, defaults, and base interfaces. The bytecode
VM uses the same interpreter operation and runtime value; it does not have a
separate dictionary representation.

Method dictionaries are already cached. The cache key contains the concrete
type descriptor, interface name, and interface arguments. Cache entries retain
a private map; later interface values receive it as `SharedMethods`, while
their local overlay and bound-method memo remain independent. Method/impl/type
registration invalidates the cache, so new resolution cannot reuse an old
visible-impl selection. Specialized host values such as iterators and futures
may supply the same dictionary contract through their runtime-native method
providers.

Compiled no-bootstrap execution may construct an interpreter-facing interface
value at an explicit boundary. Its resolver supplies compiled native callables
where an interpreter dictionary is required; that boundary accommodation does
not make the surrounding static compiled path dictionary-based.

## Compiled static interface paths

When the compiler can prove a fully bound object-safe interface shape, it
emits a generated `__able_iface_*` Go interface carrier, a concrete adapter,
and direct Go method calls. Static parameters, returns, typed locals, fields,
and direct calls stay on that carrier. Conversion helpers such as
`*_to_runtime_value` and `*_from_value` exist only at explicit runtime/dynamic
edges, callbacks, or other required ABI boundaries.

For an exact `Self` return, the generated Go method returns the same
`__able_iface_*` carrier. The concrete adapter invokes the selected compiled
implementation on its native receiver and wraps the concrete result directly
back into that carrier. No `runtime.Value`, interpreter method call, or
consumer-scope implementation lookup is part of this static path.

Do not replace those direct carriers with `runtime.InterfaceValue`, a generic
dictionary, a `runtime.Value`, or a named-container-specific lowering rule.
Conversely, do not treat a dynamic runtime value as proof that it can stay on a
static carrier without a checked conversion.

## Verification and performance rule

The focused interpreter cache tests cover shared-dictionary reuse,
per-value isolation, invalidation, bound-method memoization, and allocation
behavior. The focused compiler interface and dynamic-helper tests cover direct
static carriers and explicit dynamic helpers.

The historical proposal to add a dictionary cache is complete; there is no
selected interface-dispatch performance change. A future change must preserve
the cast-site semantics and show the same material leaf in at least three
unlike verified applications under the project performance gate. It must not
be justified by one interface, one stdlib container, or one benchmark shape.
