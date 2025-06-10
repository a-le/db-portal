const ConnInfos = {
    response: null,
    database: null,
    schema: null,
    user: null,
    version: null,
    reset: () => {
        ConnInfos.response = null;

        ConnInfos.hostname = null;
        ConnInfos.port = null;
        ConnInfos.database = null;
        ConnInfos.schema = null;
        ConnInfos.user = null;
        ConnInfos.version = null;
    },
    get: () => {
        if (App.conn === "") {
            ConnInfos.reset();
            return;
        }
        return m.request({
            method: "GET",
            url: "/api/command/:conn/:schema/conn-infos",
            credentials: "include",
            headers: getRequestHeaders(),
            extract: getRequestExtract(),
            params: { conn: App.conn, schema: App.schema },
        })
        .then((result) => {
            ConnInfos.response = result;
            if (result.DBerror)
                return;

            let i = 0;
            ConnInfos.hostname = !result.rows[0][i] ? null : { "key": result.cols[i], "value": result.rows[0][i] }; i++;
            ConnInfos.port = !result.rows[0][i] ? null : { "key": result.cols[i], "value": result.rows[0][i] }; i++;
            ConnInfos.database = !result.rows[0][i] ? null : { "key": result.cols[i], "value": result.rows[0][i] }; i++;
            ConnInfos.schema = !result.rows[0][i] ? null : { "key": result.cols[i], "value": result.rows[0][i] }; i++;
            ConnInfos.user = !result.rows[0][i] ? null : { "key": result.cols[i], "value": result.rows[0][i] }; i++;
            ConnInfos.version = !result.rows[0][i] ? null : { "key": result.cols[i], "value": result.rows[0][i] }; i++;
        })
    },
    view: () => {
        return [
            !ConnInfos.response ? null : [
                ConnInfos.response.DBerror ? m("div.text-warning", ConnInfos.response.DBerror) : [
                    !ConnInfos.database ? null :
                        m("div.font-sm.mr-20", 
                            m("span", ConnInfos.database.key + ": "),
                            m("span.info", { title: ConnInfos.database.value }, ConnInfos.database.value.split(/[/\\]/).pop())
                        ),
                    !ConnInfos.schema ? null :
                        m("div.font-sm.mr-20", 
                            m("span", ConnInfos.schema.key + ": "),
                            m("span.info", ConnInfos.schema.value)
                        ),
                    !ConnInfos.user ? null :
                        m("div.font-sm.mr-20", 
                            m("span", ConnInfos.user.key + ": "),
                            m("span.info", ConnInfos.user.value)
                        ),
                    !ConnInfos.hostname ? null :
                        m("div.font-sm.mr-20",
                            m("span", ConnInfos.hostname.key + ": "),
                            m("span.info", ConnInfos.hostname.value + (ConnInfos.port.value ? ":" + ConnInfos.port.value : ""))
                        ),
                    !ConnInfos.version ? null :
                        m("div.font-sm.mr-20.no-overflow", { style: "max-width: 220px;" },
                            m("span", ConnInfos.version.key + ": "),
                            m("span.info", { title: ConnInfos.version.value }, ConnInfos.version.value)
                        )
                ]
            ]
        ];
    }
};