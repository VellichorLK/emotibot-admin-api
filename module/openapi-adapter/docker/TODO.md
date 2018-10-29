# TODO

* 目前 `docker-compose config` 的 warning 检查只会检查是否有该传的环境变数没有传。原先是打算透过将 `docker-compose config` 的 `stderr` pipe 至 `grep` 过滤是否有 __WARNING__ 的字串出现，但在 trace 过 docker-compose 的 source codes 后发现 docker-compose 会检查输出是否为 TTY (console)，如果是的话才会加上 __WARNING__ 的 prefix 字串，没有的话就不会，因此无法透过 pipe 给 `grep` 来检查是否有 warning message。

    ```
    https://github.com/docker/compose/blob/master/compose/cli/main.py (Line 145):

    def setup_console_handler(handler, verbose, noansi=False, level=None):
        if handler.stream.isatty() and noansi is False:
            format_class = ConsoleWarningFormatter
        else:
            format_class = logging.Formatter
    ```

    因此目前先只检查是否有该传的环境变数没有传 (透过 pipe 至 `grep` 檢查 __Defaulting to a blank string__ 警示讯息是否有出现)，希望能有更完善的 warning 检查方法。

    详见 `utils.sh: valid_config()`
