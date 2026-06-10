# Channel Rollup

Channel Rollup reads a deterministic 16,384-word prefix of the checked-in
ENABLE word list, sends records through a buffered producer/worker channel
pipeline, and reduces the selected weighted scores. It measures ordinary file,
String, channel, `spawn`, and result-reduction behavior without a
benchmark-specific API or container rule.
