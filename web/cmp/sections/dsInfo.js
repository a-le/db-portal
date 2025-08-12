const DSInfoSection = {
    response: null,
    database: null,
    schema: null,
    user: null,
    version: null,
    error: null,
    reset: () => {
        DSInfoSection.response = null;
        DSInfoSection.error = null;

        DSInfoSection.hostname = null;
        DSInfoSection.port = null;
        DSInfoSection.database = null;
        DSInfoSection.schema = null;
        DSInfoSection.user = null;
        DSInfoSection.version = null;
    },
    get: () => {
        if (QueryPage.dsName === "") {
            DSInfoSection.reset();
            return;
        }

        let url, params;
        params = { dsName: QueryPage.dsName };
        if (QueryPage.schema !== "") {
            url = "/api/command/:dsName/:schema/ds-info";
            params.schema = QueryPage.schema;
        } else {
            url = "/api/command/:dsName/ds-info";
        }

        return m.request({
            method: "GET",
            url,
            headers: App.getAuthHeaders(),
            params,
        })
            .then((r) => {
                DSInfoSection.response = r.data;
                if (r.error)
                    return;

                let i = 0;
                DSInfoSection.hostname = !r.data.rows[0][i] ? null : { "key": r.data.cols[i], "value": r.data.rows[0][i] }; i++;
                DSInfoSection.port = !r.data.rows[0][i] ? null : { "key": r.data.cols[i], "value": r.data.rows[0][i] }; i++;
                DSInfoSection.database = !r.data.rows[0][i] ? null : { "key": r.data.cols[i], "value": r.data.rows[0][i] }; i++;
                DSInfoSection.schema = !r.data.rows[0][i] ? null : { "key": r.data.cols[i], "value": r.data.rows[0][i] }; i++;
                DSInfoSection.user = !r.data.rows[0][i] ? null : { "key": r.data.cols[i], "value": r.data.rows[0][i] }; i++;
                DSInfoSection.version = !r.data.rows[0][i] ? null : { "key": r.data.cols[i], "value": r.data.rows[0][i] }; i++;
            })
    },
    view: () => {
        return [
            !DSInfoSection.response ? null : [
                DSInfoSection.error ? m("div.text-warning", DSInfoSection.error) : [
                    !DSInfoSection.database ? null :
                        m("div.font-sm.mr-20",
                            m("span", DSInfoSection.database.key + ": "),
                            m("span.info", { title: DSInfoSection.database.value }, DSInfoSection.database.value.split(/[/\\]/).pop())
                        ),
                    !DSInfoSection.schema ? null :
                        m("div.font-sm.mr-20",
                            m("span", DSInfoSection.schema.key + ": "),
                            m("span.info", DSInfoSection.schema.value)
                        ),
                    !DSInfoSection.user ? null :
                        m("div.font-sm.mr-20",
                            m("span", DSInfoSection.user.key + ": "),
                            m("span.info", DSInfoSection.user.value)
                        ),
                    !DSInfoSection.hostname ? null :
                        m("div.font-sm.mr-20",
                            m("span", DSInfoSection.hostname.key + ": "),
                            m("span.info", DSInfoSection.hostname.value + (DSInfoSection.port.value ? ":" + DSInfoSection.port.value : ""))
                        ),
                    !DSInfoSection.version ? null :
                        m("div.font-sm.mr-20.no-overflow", { style: "max-width: 220px;" },
                            m("span", DSInfoSection.version.key + ": "),
                            m("span.info", { title: DSInfoSection.version.value }, DSInfoSection.version.value)
                        )
                ]
            ]
        ];
    }
};