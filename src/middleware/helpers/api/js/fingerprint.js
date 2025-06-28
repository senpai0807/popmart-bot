const { asyncData } = require("./vm.cjs");

asyncData()
    .then(result => {
        console.log(result);
    })
    .catch(err => {
        console.error("JS Error:", err);
        process.exit(1);
    });