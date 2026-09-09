class User(val name: String)

fun getUserName(user: User?): String {
    return user!!.name
}
