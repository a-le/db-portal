function DataSourceInput() {
    return {
        dsNames: [],
        error: "",
        connecting: false,
        getDataSources: function () {
            m.request({
                method: "GET",
                url: "/api/users/:username/data-sources",
                params: { username: App.getUsername() },
                headers: App.getAuthHeaders(),
            }).then((response) => {
                this.dsNames = response.data;
            });
        },
        testDataSource: function (dsName, onSuccess) {
            this.connecting = true
            this.error = ""
            m.request({
                method: "GET",
                url: "/api/users/:username/data-sources/:dsName/test",
                params: { username: App.getUsername(), dsName: dsName },
                headers: App.getAuthHeaders(),
            }).then((response) => {
                this.connecting = false;                    
                if (onSuccess) onSuccess(dsName, response)
            }).catch((e) => {
                this.connecting = false; 
                this.error = e.response.error
            })
        },
        oninit: function () {
            this.getDataSources()
        },
        view: function (vnode) {
            if (!this.dsNames.length)
                return null;

            const { onConnect, onChange, namePrefix, value = "" } = vnode.attrs || {};
            const name = namePrefix ? `${namePrefix}[dsName]` : "dsName";
            const options = toSelectOptions(this.dsNames, "select data source…", "name", "name", "vendor");
            return [
                m(SelectInput, {
                    name,
                    value,
                    options,
                    onchange: (e) => {
                        this.error = "";
                        this.connecting = false;
                        const val = e.target.value;
                        if (onChange) onChange(val);
                        if (val === "") return;
                        this.testDataSource(val, onConnect);
                    },
                    onfocus: () => {
                        this.getDataSources();
                    }
                }),
                this.error && m("span.error.ml-10", this.error),
                this.connecting && m(WaitingAnimation, { text: "connecting", class: "ml-10" })
            ];
        }
    };
}
