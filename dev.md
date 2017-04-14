## Dev Notes
This wraps ipc and message together to create a smart router for ipc while
leaving ipc simple

When sending a message, the Id gets wiped, could be a problem is something tries
to look at the ID after sending.

It might make more sense to have Send() and SendQuery(callback) rather than
having Query as a constructor. Also, maybe put service in the contructor.