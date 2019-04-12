from telegram.ext.commandhandler import CommandHandler
from telegram.ext.updater import Updater


class TelegramBot(Updater):
    __registry = {}

    @property
    def registered_commands(self):
        return self.__registry

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

        methods = [getattr(self, name) for name in dir(self) if not name.startswith('_')]
        commands = filter(lambda fn: getattr(fn, 'bot_command', False), methods)

        for command in commands:
            self.__registry[command.name] = command.description
            self.dispatcher.add_handler(CommandHandler(command.name, command))

    def run(self):
        self.start_polling()
        self.idle()
