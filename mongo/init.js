db = db.getSiblingDB("companies")
if (db.getUser("tuser") === null) {
    db.createUser({
        user: "tuser",
        pwd: "tpass",
        roles: [
            {
                role: "readWrite",
                db: "companies"
            },
        ]
    })
}