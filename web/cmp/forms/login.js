function LoginForm() {
    return {
        error: "",
        login: function (username, password) {
            this.error = "";
            if (!username || !password) {
                this.error = "Please enter both username and password.";
                return;
            }
            m.request({
                method: "POST",
                url: "/api/auth/login",
                body: { username, password },
            }).then((resp) => {
                if (!resp.token || resp.error) {
                    this.error = resp.error || "Login failed";
                } else {
                    App.login(resp.token);
                }
            }).catch((e) => {
                this.error = e.response.error || "Login failed";
            });
        },
        view: function () {
            return m("div", {
                style: {
                    position: "fixed",
                    top: 0,
                    left: 0,
                    width: "100vw",
                    height: "100vh",
                    background: "rgba(0,0,0,0.6)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    zIndex: 99
                }
            }, [
                m("form", {
                    //autocomplete: "off",
                    onsubmit: (e) => {
                        e.preventDefault()
                        usernmame = e.target.elements["username"].value
                        password = e.target.elements["password"].value
                        this.login(usernmame, password)
                    }
                }, [
                    m("fieldset", { style: "background-color: var(--primary-bg);" },
                        m("legend", "login"),
                        m("div", [
                            m("label", { for: "username" }, "username"),
                            m("input[type=text]", {
                                id: "username",
                                //autocomplete: "off",
                                required: 1,
                                oncreate: vnode => vnode.dom.focus()
                            })
                        ]),
                        m("div", [
                            m("label", { for: "password" }, "password"),
                            m("input[type=password]", {
                                id: "password",
                                required: 1,
                                autocomplete: "new-password",
                            })
                        ]),
                        m("button[type=submit].mt-15", "login"),
                        m("div.error", this.error),
                    ),
                ])
            ]);
        }
    };
}