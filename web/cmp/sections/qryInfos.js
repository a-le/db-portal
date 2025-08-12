const QryInfosSection = {
    rowsAffected: "",
    duration: "",
    truncated: false,
    clockResolution: null,
    oninit: () => {
        return m.request({
            method: "GET",
            url: "/api/clock-resolution",
            headers: App.getAuthHeaders(),
        })
        .then(function (result) {
            QryInfosSection.clockResolution = Math.ceil(result.data/1e+6) + " ms"; // nanoseconds to milliseconds
        })
    },
    reset: () => {
        QryInfosSection.rowsAffected = "";
        QryInfosSection.duration = "";
        QryInfosSection.truncated = false;
    },
    view: () => {
        QryInfosSection.rowsAffected = !QryForm.respData || QryForm.respData.DBerror !== "" || QryForm.respData.stmtType === "query" ? "-" : QryForm.respData.rowsAffected;
        QryInfosSection.rowsReturned = !QryForm.respData || QryForm.respData.DBerror !== "" || QryForm.respData.stmtType === "non-query" ? "-" : QryForm.respData.rowsReturned;
        QryInfosSection.duration = !QryForm.respData ? "-" : (QryForm.respData.duration == 0 ? "<" + QryInfosSection.clockResolution : QryForm.respData.duration) + " ms";
        QryInfosSection.truncated = !QryForm.respData ? false : QryForm.respData.truncated;
        return [ 
            !QryInfosSection.duration ? null : [
                m("div.tab-addon.font-sm.mr-20", {style: "min-width: 130px; "}, "rows returned: ", 
                    m('span.info', {
                        class: QryInfosSection.truncated ? "text-warning" : "",
                        title: QryInfosSection.truncated ? "The result is truncated. Use export to get the full result." : "",
                    }, QryInfosSection.rowsReturned)
                ),
                m("div.tab-addon.font-sm.mr-20", {style: "min-width: 130px; "}, "rows affected: ", m("span.info", QryInfosSection.rowsAffected)),
                m("div.tab-addon.font-sm.mr-20", "duration: ", m("span.info", QryInfosSection.duration)),
            ]
        ];
    }
};