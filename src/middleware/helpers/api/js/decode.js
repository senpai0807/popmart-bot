const { asyncDecData } = require('./vm.cjs');

const inputJson = process.argv[2];

let input;
try {
    input = JSON.parse(inputJson);
} catch (err) {
    console.error("Invalid JSON input");
    process.exit(1);
}

asyncDecData(input)
    .then(result => {
        console.log(JSON.stringify(result));
    })
    .catch(err => {
        console.error("Execution failed:", err);
        process.exit(1);
    });