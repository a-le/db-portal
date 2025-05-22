const LogOut = {
    view: () => {
        return [
            m("button[type=button].mb-5", {
                title: "logout " + username,
                onclick: function () {
                    sessionStorage.removeItem("token");
                    m.request({
                        method: "GET",
                        url: "/logout",
                        user: "thisIsForLogout",
                    }).catch(function (e) {
                        if (e.code === 401) {
                            console.log("User logged out.");
                            m.route.set("/");
                            window.location.reload();
                        }
                    })
                }
            }, "logout"),
        ]
    }
}


const ThemeSwitch = {
    view: () => {
        return m("div.toggle-switch", { title: "Light/Dark Mode" },
            m("label.switch",
                m("input", {
                    type: "checkbox", name: "lightswitch",
                    oncreate: function (vnode) {
                        vnode.dom.checked = isLightTheme(App.theme);
                    },
                    onclick: function (e) {
                        App.theme = (e.target.checked ? "light-mode" : "dark-mode");
                        localStorage.setItem("theme", App.theme);
                    }
                }),
                m("span.slider")
            )
        );
    }
}

// Resize panel - code was mostly AI generated
const GridResize = (gridSelector, gutterSelector, areaSelector, gridTemplate, minSize = 100, isColumn = false) => {
    const grid = document.querySelector(gridSelector);
    const gutter = document.querySelector(gutterSelector);
    const area = grid.querySelector(areaSelector);

    let isDragging = false;
    let startPosition = 0;
    let startAreaSize = 0;
    let originalSize = 0; // Store the original size for resetting on double-click

    // Helper function to add or remove the no-select class
    function setNoSelect(state) {
        state ? document.body.classList.add('no-select') : document.body.classList.remove('no-select');
    }

    const onMouseDown = (e) => {
        isDragging = true;
        startPosition = isColumn ? e.clientX : e.clientY;
        startAreaSize = isColumn ? area.offsetWidth : area.offsetHeight;
        document.body.style.cursor = isColumn ? 'col-resize' : 'row-resize';
        setNoSelect(true);
    };

    const onMouseMove = (e) => {
        if (!isDragging) return;
        const delta = isColumn ? (e.clientX - startPosition) : (e.clientY - startPosition);
        let newAreaSize = startAreaSize + delta;

        // Ensure the new size is not less than the minimum size
        newAreaSize = Math.max(newAreaSize, minSize);

        const currentSizes = grid.style[isColumn ? 'gridTemplateColumns' : 'gridTemplateRows'].split(' ');
        currentSizes[0] = `${newAreaSize}px`;
        grid.style[isColumn ? 'gridTemplateColumns' : 'gridTemplateRows'] = currentSizes.join(' ');
    };

    const onMouseUp = () => {
        isDragging = false;
        document.body.style.cursor = '';
        setNoSelect(false);
    };

    // Handle double-click to reset size to the original size
    const onDoubleClick = () => {
        const currentSizes = grid.style[isColumn ? 'gridTemplateColumns' : 'gridTemplateRows'].split(' ');
        currentSizes[0] = `${originalSize}px`; // Reset to the original size
        grid.style[isColumn ? 'gridTemplateColumns' : 'gridTemplateRows'] = currentSizes.join(' ');
    };

    return {
        init: () => {
            if (isColumn) {
                grid.style.gridTemplateColumns = gridTemplate;
                originalSize = area.offsetWidth; // Save the original size
            } else {
                grid.style.gridTemplateRows = gridTemplate;
                originalSize = area.offsetHeight; // Save the original size
            }
            gutter.addEventListener('mousedown', onMouseDown);
            gutter.addEventListener('dblclick', onDoubleClick); // Add double-click event
            document.addEventListener('mousemove', onMouseMove);
            document.addEventListener('mouseup', onMouseUp);
        },
        destroy: () => {
            // Remove event listeners
            gutter.removeEventListener('mousedown', onMouseDown);
            gutter.removeEventListener('dblclick', onDoubleClick); // Remove double-click event
            document.removeEventListener('mousemove', onMouseMove);
            document.removeEventListener('mouseup', onMouseUp);
        }
    };
};

class TabState {
    constructor(defaultTab) {
        this.currentTab = defaultTab;
    }
    set(tab) {
        this.currentTab = tab;
    }
    get() {
        return this.currentTab;
    }
    is(tab) {
        return (this.currentTab === tab);
    }
    selectedClass(tab) {
        return (this.currentTab === tab) ? "selected" : "";
    }
    displayClass(tab) {
        return (this.currentTab === tab) ? "" : "display-none";
    }
}

class TableDim {
    constructor() {
        this.rows = [[]];
        this.charWidth = 6.5;
        this.availableWidth = 0;
        this.tdPadding = 12;
        this.colWidths = new Map();
        this.totalWidth = 0;
    }

    setRows(newRows) {
        this.rows = newRows;
        return this;
    }

    setCharWidth(newCharWidth) {
        this.charWidth = newCharWidth;
        return this;
    }

    setAvailableWidth(newAvailableWidth) {
        this.availableWidth = newAvailableWidth;
        return this;
    }

    setTdPadding(newTdPadding) {
        this.tdPadding = newTdPadding;
        return this;
    }

    getColWidths() {
        return this.colWidths;
    }

    getColWidth(colIdx) {
        return Math.floor(this.colWidths.get(colIdx));
    }

    getTotalWidth() {
        return this.totalWidth;
    }

    calc() {
        const maxColLengths = new Map();
        this.colWidths = new Map();
        this.totalWidth = 0;

        this.rows.forEach(row => {
            row.forEach((col, idx) => {
                var len = 0;
                if ( col == null ) len = 4;
                else if ( typeof col == "object" ) len = JSON.stringify(col).length;
                else len = String(col).length;
                len = Math.max(len, 2);
                if (!maxColLengths.has(idx) || len > maxColLengths.get(idx)) {
                    maxColLengths.set(idx, len);
                }
            });
        });

        let remainingWidth = this.availableWidth;

        outerLoop: while (maxColLengths.size > 0) {
            const evenlyDistributedWidth = Math.floor(remainingWidth / maxColLengths.size);

            for (const [idx, maxLen] of maxColLengths) {
                const idealWidth = maxLen * this.charWidth + this.tdPadding;

                if (idealWidth <= evenlyDistributedWidth) {
                    this.colWidths.set(idx, idealWidth);
                    remainingWidth -= idealWidth;
                    maxColLengths.delete(idx);
                    continue outerLoop;
                }
            }

            for (const [idx] of maxColLengths) {
                const evenlyWidth = Math.floor(remainingWidth / maxColLengths.size);
                this.colWidths.set(idx, evenlyWidth);
                remainingWidth -= evenlyWidth;
                maxColLengths.delete(idx);
            }
            break;
        }

        if (remainingWidth > this.colWidths.size) {
            let usedWidth = 0;
            for (const [, v] of this.colWidths) {
                usedWidth += v;
            }

            const addWidth = Math.floor(Math.min(remainingWidth / this.colWidths.size, usedWidth * 0.2 / this.colWidths.size));

            for (const [k, v] of this.colWidths) {
                this.colWidths.set(k, v + addWidth);
            }
        }

        for (const [, v] of this.colWidths) {
            this.totalWidth += v;
        }

        return this;
    }
}


// Mithril component that accepts rows as an argument
// rows should be an array of arrays with 2 cols
const OptGroup = {
    // `vnode.attrs.rows` is expected to be the input array
    view: function (vnode) {
        let currentGroup = null;
        let elements = [];

        // Sort by second column, then by first column
        vnode.attrs.rows.sort((a, b) => {
            // First compare by the second column
            if (a[1] < b[1]) return -1;
            if (a[1] > b[1]) return 1;

            // If second columns are equal, compare by the first column
            if (a[0] < b[0]) return -1;
            if (a[0] > b[0]) return 1;

            return 0; // If both are equal
        });

        vnode.attrs.rows.forEach(function (row) {
            let groupLabel = row[1]; // Group by row[1] 
            let optionElement = m("option", { value: row[0] }, row[0]);

            // Check if we need to start a new optgroup
            if (currentGroup !== groupLabel) {
                currentGroup = groupLabel;

                // Add new optgroup with the option inside
                elements.push(m("optgroup", { label: currentGroup }, [optionElement]));
            } else {
                // Add option to the last optgroup
                elements[elements.length - 1].children.push(optionElement);
            }
        });

        return elements;
    }
};


const Cell = {
    classTar(type) {
        t = type.toLowerCase();
        return ["string", "name", "text", "unknown"].includes(t) || t.includes("char") ? "" : "tar";
    },
    classCell(val) {
        if (val === null) return "cell-null";
        if (val === true) return "cell-true";
        if (val === false) return "cell-false";
        return "";
    },
    displayValue(val) {
        if (val === null) return "null";
        if (val === true) return "true";
        if (val === false) return "false";
        if (typeof val == "object") return JSON.stringify(val);
        return val;
    },
    setTitleFromInnerText(e) {
        if ( e.currentTarget.title === "" ) e.currentTarget.title = e.currentTarget.innerText;
    },
    view: (vnode) => {
        var displayVal = Cell.displayValue(vnode.attrs.val);
        return m("td", m("div", {
            onmouseover: (e) => {
                Cell.setTitleFromInnerText(e);
            },
            class: [Cell.classTar(vnode.attrs.type), Cell.classCell(displayVal)].join(" "),
        }, Cell.displayValue(displayVal)
        ))
    }
}

const isLightTheme = (theme) => {
    return theme.toLowerCase().includes('light');
}



function WaitingAnimation(vnode) {
    const animationState = {
        baseText: vnode.attrs.text || "processing",
        repeatChar: vnode.attrs.repeatChar || ".",
        dots: 0,
        displayText: ""
    };

    // Set up the interval to update dots and trigger a redraw
    const startAnimation = () => {
        animationState.interval = setInterval(() => {
            animationState.dots = (animationState.dots + 1) % 4;
            animationState.displayText = animationState.baseText + animationState.repeatChar.repeat(animationState.dots);
            m.redraw(); // Trigger a redraw to update the view
        }, 500);
    };

    // Clear the interval when component is removed
    const stopAnimation = () => {
        clearInterval(animationState.interval);
    };

    return {
        oncreate: startAnimation,
        onremove: stopAnimation,
        view: () => m("span", {class: vnode.attrs.class}, animationState.displayText || animationState.baseText)
    };
}


const getRequestHeaders = () => {
    return sessionStorage.getItem("token") !== null ? {
        "Authorization-Jwt": "Bearer " + sessionStorage.getItem("token")
    } : {}
}

const getRequestExtract = () => {
    return function (xhr) {
        const token = xhr.getResponseHeader("Authorization-Jwt"); // Assuming server sends "Authorization: Bearer <token>" in headers
        if (token) {
            sessionStorage.setItem("token", token.split(" ")[1]); // Store the token without the "Bearer " prefix
        }
        return JSON.parse(xhr.responseText); // Continue to parse response as JSON
    }
}

const DownloadIcon = {
  view: ({ attrs }) => {
    const {
      size = 18,
      color = "currentColor",
      className = "",
      style = "",
      ...rest
    } = attrs;
    
    // Ensure vertical-align is always applied to the outer <span>
    const spanStyle = `vertical-align: middle;${style ? " " + style : ""}`;

    const svg = `
      <svg xmlns="http://www.w3.org/2000/svg"
           viewBox="0 0 24 24"
           width="${size}"
           height="${size}"
           fill="${color}"
           class="${className}"
           style="${style}">
        <path d="M13 12H16L12 16L8 12H11V8H13V12ZM15 4H5V20H19V8H15V4ZM3 2.9918C3 2.44405 3.44749 2 3.9985 2H16L20.9997 7L21 20.9925C21 21.5489 20.5551 22 20.0066 22H3.9934C3.44476 22 3 21.5447 3 21.0082V2.9918Z"/>
      </svg>
    `;

    return m("span", {rest, style: spanStyle}, m.trust(svg));
  }
};
