def bot_command(name, description):
    def bot_command_decorator(func):
        func.bot_command = True
        func.name = name
        func.description = description
        return func
    return bot_command_decorator


def admin_required(func):
    def wrapper(self, bot, update):
        from_user = update.message.from_user.username
        if from_user in self.admins:
            return func(self, bot, update)
        update.message.reply_text(f'You don\'t have access to run this command')
    return wrapper
