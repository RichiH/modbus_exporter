## 0.5.0 / 2026-10-16

* [CHANGE] Switch logging to log/slog. Log output format changes from `ts=… caller=… level=info` to `time=… source=… level=INFO` (#99)
* [CHANGE] Require Go 1.26 or newer to build (#98)
* [FEATURE] Expose device specific special values via `specialValues` as `modbus_special_value_caught` (#32)
* [FEATURE] Allow specifying multiple config files (#64)
* [FEATURE] Add `sleepAfterConnect` workaround (#65)
* [FEATURE] Retry failed scrapes, configurable via `scrapeErrorRetryCount` and `scrapeErrorWait` workarounds (#66)
* [FEATURE] Add `scrapeInterludeWait` workaround to wait between reads (#83)
* [ENHANCEMENT] Allow `bitOffset` up to 15 for bool on holding and input registers (#72)
* [ENHANCEMENT] Clarify `InsufficientRegistersError` messages (#56)
* [ENHANCEMENT] Update dependencies, including client_golang 1.25, exporter-toolkit 0.20, golang.org/x/crypto 0.58 (#82, #86, #89, #92, #99)
* [BUGFIX] Fix panic when reading bool from coils and discrete inputs (#97)
* [BUGFIX] Fix concurrent map access fatal error on parallel scrapes (#75)
* [DOCS] Add systemd instructions and clarify getting started (#67, #68, #70)

### New Contributors

* @equinox0815 made their first contribution in #56
* @Hafpaf made their first contribution in #67
* @hikhvar made their first contribution in #65
* @omnistat made their first contribution in #64
* @zzzorgo made their first contribution in #72
* @giantlock made their first contribution in #75
* @sc-alex made their first contribution in #83

## 0.4.1 / 2023-03-14

NOOP: Trigger new release

## 0.4.0

* [FEATURE] Add counter for errors and change the entry point of exporter metrics (#7)
* [FEATURE] New easy-to-write register addresses (#23)
* [FEATURE] Implement exporter-toolkit (#42)
* [ENHANCEMENT] Adjusted the typo fix (#24)
* [ENHANCEMENT] Cleanup uncessary metrics (#21)
* [ENHANCEMENT] Improve and clarify example (#26)
* [ENHANCEMENT] Set metrics_path in prometheus.yml (#31)
* [ENHANCEMENT] Remove stale license information(#33)
* [ENHANCEMENT] README: mention that sub_module is a required parameter and target needs to include the port (#36)
* [ENHANCEMENT] Fix README Diagram (#45)
* [ENHANCEMENT] Fix another diagram typo (#48)

## 0.3.0-rc.0

* [CHANGE] Remove serial connection option (#17)
* [CHANGE] Move metrics port to :9602 (#20)
* [CHANGE] Configure modbus unit aka. sub target in prometheus.yml (#21)
* [CHANGE] Specify modbus function by number not string (#22)
