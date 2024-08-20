package fix

type ExecutionReportHandler func(o *Order)

func (c *Client) SubscribeToExecutionReport(listener ExecutionReportHandler) {
	c.emitter.On(ExecutionReportTopic, listener)
}
