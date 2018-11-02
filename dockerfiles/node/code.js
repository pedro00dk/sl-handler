module.exports.helloWorld = (req, res) => {
    console.log("hello world")
    res.send(req.method)
}

module.exports.hello = (req, res) => {
    console.log("hello")
    res.send("hello")
}
