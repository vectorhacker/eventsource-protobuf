// Code generated by eventsource-protobuf. DO NOT EDIT.
// source: commands.proto

package main


// AggregateID implements the eventsource.Command interface for RegisterUser
func (c *RegisterUser) AggregateID() string {
	return c.ID
}


// AggregateID implements the eventsource.Command interface for ResetPassword
func (c *ResetPassword) AggregateID() string {
	return c.Id
}
