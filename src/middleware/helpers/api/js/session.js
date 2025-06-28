const { asyncSign } = require("./vm.cjs");

const [tokenId, path, body] = process.argv.slice(2);

asyncSign(tokenId, path, body).then(result => {
    console.log(JSON.stringify(result));
}).catch(err => {
    console.error("Execution failed:", err);
    process.exit(1);
});