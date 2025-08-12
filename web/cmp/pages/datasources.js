function DatasourcesPage() {
    return {
        registeredDSs: [],
        notRegisteredDSs: [],
        users: [],
        vendors: [],

        usernameInput: "",
        vendorInput: "",
        locationInput: "",

        postUserDatasourcesError: "",
        postUsersError: "",
        addDSError: "",
        testDataSourceResult: "",

        getUsersDatasources: function (username) {
            m.request({
                method: "GET",
                url: "/api/users/:username/data-sources",
                params: { username },
                headers: App.getAuthHeaders(),
            }).then((response) => {
                this.registeredDSs = response.data || [];
            });
        },
        getUsersAvailableDatasources: function (username) {
            m.request({
                method: "GET",
                url: "/api/users/:username/available-data-sources",
                params: { username },
                headers: App.getAuthHeaders(),
            }).then((response) => {
                this.notRegisteredDSs = response.data || [];
            });
        },
        getUsers: function () {
            m.request({
                method: "GET",
                url: "/api/users",
                headers: App.getAuthHeaders(),
            }).then((response) => {
                this.users = response.data || [];
            });
        },
        getVendors: function () {
            m.request({
                method: "GET",
                url: "/api/vendors",
                headers: App.getAuthHeaders(),
            }).then((response) => {
                this.vendors = response.data || [];
            });
        },
        postDatasources: function (name, vendor, location) {
            this.addDSError = "";
            return m.request({
                method: "POST",
                url: "/api/data-sources",
                headers: App.getAuthHeaders(),
                body: { name, vendor, location }
            }).then(() => {
                this.getUsersAvailableDatasources(this.usernameInput);
            }).catch((e) => {
                this.addDSError = e.response.error;
                throw e;
            });
        },
        postUserDatasources: function (username, dsname) {
            this.postUserDatasourcesError = "";
            return m.request({
                method: "POST",
                url: "/api/users/:username/data-sources/:dsname",
                headers: App.getAuthHeaders(),
                params: { username, dsname }
            }).then((response) => {
                this.getUsersDatasources(this.usernameInput);
                this.getUsersAvailableDatasources(this.usernameInput);
            }).catch((e) => {
                this.postUserDatasourcesError = e.response.error;
                throw e;
            });
        },
        deleteUserDatasources: function (username, dsname) {
            this.postUserDatasourcesError = "";
            return m.request({
                method: "DELETE",
                url: "/api/users/:username/data-sources/:dsname",
                headers: App.getAuthHeaders(),
                params: { username, dsname }
            }).then(() => {
                this.getUsersDatasources(this.usernameInput);
                this.getUsersAvailableDatasources(this.usernameInput);
            }).catch((e) => {
                this.postUserDatasourcesError = e.response.error;
                throw e;
            });
        },
        postUsers: function (username, isadmin, password) {
            this.postUsersError = "";
            return m.request({
                method: "POST",
                url: "/api/users",
                headers: App.getAuthHeaders(),
                body: { username, isadmin, password }
            }).then(() => {
                this.getUsers();
            }).catch((e) => {
                this.postUsersError = e.response.error;
                throw e;
            });
        },
        testDataSource: function (vendor, location) {
            return m.request({
                method: "POST",
                url: "/api/data-sources/test",
                headers: App.getAuthHeaders(),
                body: { vendor, location }
            }).then((resp) => {
                this.testDataSourceResult = resp.error ? resp.error : "connection test succeeded.";
            }).catch((e) => {
                this.testDataSourceResult = e.response.error;
                throw e;
            });
        },

        oninit: function () {
            this.usernameInput = App.getUsername();

            this.postUserDatasourcesError = "";
            this.postUsersError = "";
            this.addDSError = "";
            this.testDataSourceResult = "";

            this.getUsers();
            this.getVendors();
            this.getUsersDatasources(this.usernameInput);
            this.getUsersAvailableDatasources(this.usernameInput);
        },
        view: function () {
            return [
                m("form.mt-20", {
                    autocomplete: "off",
                    onchange: () => {
                        this.postUsersError = "";
                    },
                    onsubmit: (e) => {
                        e.preventDefault();
                        const username = e.target.elements["username"].value;
                        const dsname = e.target.elements["dsname"].value;
                        this.postUserDatasources(username, dsname)
                            .then(() => {
                                e.target.reset();
                            });
                    }
                },
                    m("fieldset",
                        m("legend", "allowed data sources"),
                        m("label",
                            m("span", "user: "),
                            m(SelectInput, {
                                name: "username",
                                required: 1,
                                options: toSelectOptions(this.users, false, "name", "name"),
                                value: this.usernameInput,
                                onchange: (e) => {
                                    this.usernameInput = e.target.value;
                                    this.getUsersDatasources(this.usernameInput);
                                    this.getUsersAvailableDatasources(this.usernameInput);
                                }
                            })
                        ),
                        m("table.mt-10", { style: { minWidth: "500px" } },
                            m("thead",
                                m("tr", [
                                    m("th", "vendor"),
                                    m("th", "name"),
                                    m("th", "location"),
                                    App.getIsAdmin() && m("th", "action"),
                                ])
                            ),
                            m("tbody",
                                this.registeredDSs.length === 0
                                    ? m("tr",
                                        m("td[colspan=4]", "No data sources found.")
                                    )
                                    : this.registeredDSs.map(row =>
                                        m("tr", [
                                            m(Cell, { val: row.vendor, type: "string" }),
                                            m(Cell, { val: row.name, type: "string" }),
                                            m(Cell, { val: row.location, type: "string" }),

                                            // remove data source
                                            App.getIsAdmin() && m("td.tar",
                                                m("button[type=button]", {
                                                    title: "remove",
                                                    onclick: (e) => {
                                                        this.deleteUserDatasources(this.usernameInput, row.name);
                                                    }
                                                }, m.trust("&#10006;"))
                                            )
                                        ])
                                    )
                            )
                        ),

                        // allow data source to user 
                        App.getIsAdmin() &&
                        m("div.mt-20",
                            m("label",
                                m("span", [
                                    "allow ",
                                    m("b.fake-input", this.usernameInput),
                                    " to access: "
                                ]),
                                m(SelectInput, {
                                    name: "dsname",
                                    required: 1,
                                    options: toSelectOptions(this.notRegisteredDSs, "", "name", "name", "vendor")
                                }),
                            ),
                            m("button[type=submit]", "submit"),
                            m("div", this.postUserDatasourcesError)
                        ),
                    ),
                ),

                // add a new data source form
                App.getIsAdmin() &&
                m("form.mt-30", {
                    autocomplete: "off",
                    onchange: () => {
                        this.addDSError = "";
                        this.testDataSourceResult = "";
                    },
                    onsubmit: (e) => {
                        e.preventDefault();
                        const name = e.target.elements["name"].value;
                        const vendor = e.target.elements["vendor"].value;
                        const location = e.target.elements["location"].value;
                        this.postDatasources(name, vendor, location)
                            .then(() => {
                                e.target.reset();
                            });;
                    }
                },
                    m("fieldset",
                        m("legend", "add a new data source"),
                        m("table",
                            m("tr",
                                m("td", {
                                    title: "the label of the data source",
                                }, "name:"),
                                m("td", m("input", {
                                    name: "name",
                                    required: 1,
                                    pattern: "^[a-zA-Z0-9_\\-]{1,30}$",
                                }))
                            ),
                            m("tr",
                                m("td", "DB vendor:"),
                                m("td", m(SelectInput, {
                                    name: "vendor",
                                    required: 1,
                                    options: toSelectOptions(this.vendors, "", "name", "name"),
                                    value: this.vendorInput,
                                    onchange: (e) => { this.vendorInput = e.target.value; }
                                }))
                            ),
                            m("tr",
                                m("td", {
                                    title: "the driver-specific data source name, usually consisting of at least a database name and connection information",
                                }, "location:"),
                                m("td", m("input", {
                                    name: "location",
                                    required: 1,
                                    value: this.locationInput,
                                    oninput: (e) => { this.locationInput = e.target.value; }
                                }))
                            ),
                            m("tr",
                                m("td"),
                                m("td",
                                    m("button[type=submit].mr-20", "add"),
                                    m("button[type=button].mr-10", {
                                        disabled: !this.vendorInput || !this.locationInput,
                                        onclick: () => {
                                            this.testDataSource(this.vendorInput, this.locationInput)
                                        }
                                    }, "test"),
                                )
                            )
                        ),
                        m("pre", this.testDataSourceResult),
                        m("span.error", this.addDSError)
                    )
                ),

                // add a new user form
                App.getIsAdmin() &&
                m("form.mt-30", {
                    autocomplete: "off",
                    onchange: () => {
                        this.postUsersError = "";
                    },
                    onsubmit: (e) => {
                        e.preventDefault();
                        const username = e.target.elements["name"].value;
                        const isadmin = e.target.elements["isadmin"].value;
                        const password = e.target.elements["password"].value;
                        this.postUsers(username, isadmin, password)
                            .then(() => {
                                e.target.reset();
                            });
                    }
                },
                    m("fieldset",
                        m("legend", "add a new user"),
                        m("table",
                            m("tr",
                                m("td", "name:"),
                                m("td", m("input", {
                                    name: "name",
                                    required: 1,
                                    pattern: "^[a-zA-Z0-9_\\-]{1,20}$",
                                }))
                            ),
                            m("tr",
                                m("td", "admin:"),
                                m("td", m(SelectInput, {
                                    name: "isadmin",
                                    required: 1,
                                    options: [
                                        { label: "", value: "" }, { label: "no", value: "0" }, { label: "yes", value: "1" }
                                    ]
                                }))
                            ),
                            m("tr",
                                m("td", "password:"),
                                m("td", m("input[type=password]", {
                                    name: "password",
                                    required: 1,
                                }))
                            ),
                            m("tr",
                                m("td"),
                                m("td", m("button[type=submit]", "add"))
                            )
                        ),
                        m("span.error", this.postUsersError),
                    )
                ),

            ];
        }
    }
}