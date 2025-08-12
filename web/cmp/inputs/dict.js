class DictInput {
    constructor() {
        this.selected = ""; // selected item name
        this.list = null; // API response object of list of items
        this.definition = null; // API response - DDL definition of selected item
        this.columns = null; // API response - list of table columns
    }
    /**
     * Change the selected item
     * @param {string} item - The item to select.
     * @returns {boolean} - Returns true if new selection needs an API call; false otherwise.
     */
    changeItem(item) {
        if (this.selected == item)
            return false;

        this.selected = item;

        // reset
        this.definition = null; // DDL definition of selected item
        this.columns = null; // API response object of selected item

        if (item === "")
            return false;

        return true;
    }
}