// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/dialect/sql/sqljson"
	"entgo.io/ent/schema/field"
	"github.com/DeedleFake/sips"
	"github.com/DeedleFake/sips/ent/pin"
	"github.com/DeedleFake/sips/ent/predicate"
	"github.com/DeedleFake/sips/ent/user"
)

// PinUpdate is the builder for updating Pin entities.
type PinUpdate struct {
	config
	hooks    []Hook
	mutation *PinMutation
}

// Where appends a list predicates to the PinUpdate builder.
func (pu *PinUpdate) Where(ps ...predicate.Pin) *PinUpdate {
	pu.mutation.Where(ps...)
	return pu
}

// SetUpdateTime sets the "update_time" field.
func (pu *PinUpdate) SetUpdateTime(t time.Time) *PinUpdate {
	pu.mutation.SetUpdateTime(t)
	return pu
}

// SetStatus sets the "Status" field.
func (pu *PinUpdate) SetStatus(ss sips.RequestStatus) *PinUpdate {
	pu.mutation.SetStatus(ss)
	return pu
}

// SetNillableStatus sets the "Status" field if the given value is not nil.
func (pu *PinUpdate) SetNillableStatus(ss *sips.RequestStatus) *PinUpdate {
	if ss != nil {
		pu.SetStatus(*ss)
	}
	return pu
}

// SetName sets the "Name" field.
func (pu *PinUpdate) SetName(s string) *PinUpdate {
	pu.mutation.SetName(s)
	return pu
}

// SetCID sets the "CID" field.
func (pu *PinUpdate) SetCID(s string) *PinUpdate {
	pu.mutation.SetCID(s)
	return pu
}

// SetOrigins sets the "Origins" field.
func (pu *PinUpdate) SetOrigins(s []string) *PinUpdate {
	pu.mutation.SetOrigins(s)
	return pu
}

// AppendOrigins appends s to the "Origins" field.
func (pu *PinUpdate) AppendOrigins(s []string) *PinUpdate {
	pu.mutation.AppendOrigins(s)
	return pu
}

// ClearOrigins clears the value of the "Origins" field.
func (pu *PinUpdate) ClearOrigins() *PinUpdate {
	pu.mutation.ClearOrigins()
	return pu
}

// SetUserID sets the "User" edge to the User entity by ID.
func (pu *PinUpdate) SetUserID(id int) *PinUpdate {
	pu.mutation.SetUserID(id)
	return pu
}

// SetNillableUserID sets the "User" edge to the User entity by ID if the given value is not nil.
func (pu *PinUpdate) SetNillableUserID(id *int) *PinUpdate {
	if id != nil {
		pu = pu.SetUserID(*id)
	}
	return pu
}

// SetUser sets the "User" edge to the User entity.
func (pu *PinUpdate) SetUser(u *User) *PinUpdate {
	return pu.SetUserID(u.ID)
}

// Mutation returns the PinMutation object of the builder.
func (pu *PinUpdate) Mutation() *PinMutation {
	return pu.mutation
}

// ClearUser clears the "User" edge to the User entity.
func (pu *PinUpdate) ClearUser() *PinUpdate {
	pu.mutation.ClearUser()
	return pu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (pu *PinUpdate) Save(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	pu.defaults()
	if len(pu.hooks) == 0 {
		if err = pu.check(); err != nil {
			return 0, err
		}
		affected, err = pu.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*PinMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = pu.check(); err != nil {
				return 0, err
			}
			pu.mutation = mutation
			affected, err = pu.sqlSave(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(pu.hooks) - 1; i >= 0; i-- {
			if pu.hooks[i] == nil {
				return 0, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = pu.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, pu.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// SaveX is like Save, but panics if an error occurs.
func (pu *PinUpdate) SaveX(ctx context.Context) int {
	affected, err := pu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (pu *PinUpdate) Exec(ctx context.Context) error {
	_, err := pu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pu *PinUpdate) ExecX(ctx context.Context) {
	if err := pu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (pu *PinUpdate) defaults() {
	if _, ok := pu.mutation.UpdateTime(); !ok {
		v := pin.UpdateDefaultUpdateTime()
		pu.mutation.SetUpdateTime(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pu *PinUpdate) check() error {
	if v, ok := pu.mutation.Status(); ok {
		if err := pin.StatusValidator(v); err != nil {
			return &ValidationError{Name: "Status", err: fmt.Errorf(`ent: validator failed for field "Pin.Status": %w`, err)}
		}
	}
	if v, ok := pu.mutation.Name(); ok {
		if err := pin.NameValidator(v); err != nil {
			return &ValidationError{Name: "Name", err: fmt.Errorf(`ent: validator failed for field "Pin.Name": %w`, err)}
		}
	}
	if v, ok := pu.mutation.CID(); ok {
		if err := pin.CIDValidator(v); err != nil {
			return &ValidationError{Name: "CID", err: fmt.Errorf(`ent: validator failed for field "Pin.CID": %w`, err)}
		}
	}
	return nil
}

func (pu *PinUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   pin.Table,
			Columns: pin.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: pin.FieldID,
			},
		},
	}
	if ps := pu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := pu.mutation.UpdateTime(); ok {
		_spec.SetField(pin.FieldUpdateTime, field.TypeTime, value)
	}
	if value, ok := pu.mutation.Status(); ok {
		_spec.SetField(pin.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := pu.mutation.Name(); ok {
		_spec.SetField(pin.FieldName, field.TypeString, value)
	}
	if value, ok := pu.mutation.CID(); ok {
		_spec.SetField(pin.FieldCID, field.TypeString, value)
	}
	if value, ok := pu.mutation.Origins(); ok {
		_spec.SetField(pin.FieldOrigins, field.TypeJSON, value)
	}
	if value, ok := pu.mutation.AppendedOrigins(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, pin.FieldOrigins, value)
		})
	}
	if pu.mutation.OriginsCleared() {
		_spec.ClearField(pin.FieldOrigins, field.TypeJSON)
	}
	if pu.mutation.UserCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   pin.UserTable,
			Columns: []string{pin.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: user.FieldID,
				},
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   pin.UserTable,
			Columns: []string{pin.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: user.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, pu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{pin.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	return n, nil
}

// PinUpdateOne is the builder for updating a single Pin entity.
type PinUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *PinMutation
}

// SetUpdateTime sets the "update_time" field.
func (puo *PinUpdateOne) SetUpdateTime(t time.Time) *PinUpdateOne {
	puo.mutation.SetUpdateTime(t)
	return puo
}

// SetStatus sets the "Status" field.
func (puo *PinUpdateOne) SetStatus(ss sips.RequestStatus) *PinUpdateOne {
	puo.mutation.SetStatus(ss)
	return puo
}

// SetNillableStatus sets the "Status" field if the given value is not nil.
func (puo *PinUpdateOne) SetNillableStatus(ss *sips.RequestStatus) *PinUpdateOne {
	if ss != nil {
		puo.SetStatus(*ss)
	}
	return puo
}

// SetName sets the "Name" field.
func (puo *PinUpdateOne) SetName(s string) *PinUpdateOne {
	puo.mutation.SetName(s)
	return puo
}

// SetCID sets the "CID" field.
func (puo *PinUpdateOne) SetCID(s string) *PinUpdateOne {
	puo.mutation.SetCID(s)
	return puo
}

// SetOrigins sets the "Origins" field.
func (puo *PinUpdateOne) SetOrigins(s []string) *PinUpdateOne {
	puo.mutation.SetOrigins(s)
	return puo
}

// AppendOrigins appends s to the "Origins" field.
func (puo *PinUpdateOne) AppendOrigins(s []string) *PinUpdateOne {
	puo.mutation.AppendOrigins(s)
	return puo
}

// ClearOrigins clears the value of the "Origins" field.
func (puo *PinUpdateOne) ClearOrigins() *PinUpdateOne {
	puo.mutation.ClearOrigins()
	return puo
}

// SetUserID sets the "User" edge to the User entity by ID.
func (puo *PinUpdateOne) SetUserID(id int) *PinUpdateOne {
	puo.mutation.SetUserID(id)
	return puo
}

// SetNillableUserID sets the "User" edge to the User entity by ID if the given value is not nil.
func (puo *PinUpdateOne) SetNillableUserID(id *int) *PinUpdateOne {
	if id != nil {
		puo = puo.SetUserID(*id)
	}
	return puo
}

// SetUser sets the "User" edge to the User entity.
func (puo *PinUpdateOne) SetUser(u *User) *PinUpdateOne {
	return puo.SetUserID(u.ID)
}

// Mutation returns the PinMutation object of the builder.
func (puo *PinUpdateOne) Mutation() *PinMutation {
	return puo.mutation
}

// ClearUser clears the "User" edge to the User entity.
func (puo *PinUpdateOne) ClearUser() *PinUpdateOne {
	puo.mutation.ClearUser()
	return puo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (puo *PinUpdateOne) Select(field string, fields ...string) *PinUpdateOne {
	puo.fields = append([]string{field}, fields...)
	return puo
}

// Save executes the query and returns the updated Pin entity.
func (puo *PinUpdateOne) Save(ctx context.Context) (*Pin, error) {
	var (
		err  error
		node *Pin
	)
	puo.defaults()
	if len(puo.hooks) == 0 {
		if err = puo.check(); err != nil {
			return nil, err
		}
		node, err = puo.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*PinMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = puo.check(); err != nil {
				return nil, err
			}
			puo.mutation = mutation
			node, err = puo.sqlSave(ctx)
			mutation.done = true
			return node, err
		})
		for i := len(puo.hooks) - 1; i >= 0; i-- {
			if puo.hooks[i] == nil {
				return nil, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = puo.hooks[i](mut)
		}
		v, err := mut.Mutate(ctx, puo.mutation)
		if err != nil {
			return nil, err
		}
		nv, ok := v.(*Pin)
		if !ok {
			return nil, fmt.Errorf("unexpected node type %T returned from PinMutation", v)
		}
		node = nv
	}
	return node, err
}

// SaveX is like Save, but panics if an error occurs.
func (puo *PinUpdateOne) SaveX(ctx context.Context) *Pin {
	node, err := puo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (puo *PinUpdateOne) Exec(ctx context.Context) error {
	_, err := puo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (puo *PinUpdateOne) ExecX(ctx context.Context) {
	if err := puo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (puo *PinUpdateOne) defaults() {
	if _, ok := puo.mutation.UpdateTime(); !ok {
		v := pin.UpdateDefaultUpdateTime()
		puo.mutation.SetUpdateTime(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (puo *PinUpdateOne) check() error {
	if v, ok := puo.mutation.Status(); ok {
		if err := pin.StatusValidator(v); err != nil {
			return &ValidationError{Name: "Status", err: fmt.Errorf(`ent: validator failed for field "Pin.Status": %w`, err)}
		}
	}
	if v, ok := puo.mutation.Name(); ok {
		if err := pin.NameValidator(v); err != nil {
			return &ValidationError{Name: "Name", err: fmt.Errorf(`ent: validator failed for field "Pin.Name": %w`, err)}
		}
	}
	if v, ok := puo.mutation.CID(); ok {
		if err := pin.CIDValidator(v); err != nil {
			return &ValidationError{Name: "CID", err: fmt.Errorf(`ent: validator failed for field "Pin.CID": %w`, err)}
		}
	}
	return nil
}

func (puo *PinUpdateOne) sqlSave(ctx context.Context) (_node *Pin, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   pin.Table,
			Columns: pin.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: pin.FieldID,
			},
		},
	}
	id, ok := puo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Pin.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := puo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, pin.FieldID)
		for _, f := range fields {
			if !pin.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != pin.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := puo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := puo.mutation.UpdateTime(); ok {
		_spec.SetField(pin.FieldUpdateTime, field.TypeTime, value)
	}
	if value, ok := puo.mutation.Status(); ok {
		_spec.SetField(pin.FieldStatus, field.TypeEnum, value)
	}
	if value, ok := puo.mutation.Name(); ok {
		_spec.SetField(pin.FieldName, field.TypeString, value)
	}
	if value, ok := puo.mutation.CID(); ok {
		_spec.SetField(pin.FieldCID, field.TypeString, value)
	}
	if value, ok := puo.mutation.Origins(); ok {
		_spec.SetField(pin.FieldOrigins, field.TypeJSON, value)
	}
	if value, ok := puo.mutation.AppendedOrigins(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, pin.FieldOrigins, value)
		})
	}
	if puo.mutation.OriginsCleared() {
		_spec.ClearField(pin.FieldOrigins, field.TypeJSON)
	}
	if puo.mutation.UserCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   pin.UserTable,
			Columns: []string{pin.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: user.FieldID,
				},
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   pin.UserTable,
			Columns: []string{pin.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: user.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Pin{config: puo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, puo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{pin.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	return _node, nil
}
