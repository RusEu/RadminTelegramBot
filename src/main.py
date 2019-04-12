import os

from .lib.telegram import TelegramBot
from .lib.telegram.decorators import bot_command, admin_required
from . import config


class Bot(TelegramBot):
    admins = []

    def __init__(self, *args, **kwargs):
        self.admins = kwargs.pop('admins', [])
        super().__init__(*args, **kwargs)

    @bot_command(name='help', description='List all commands')
    def help_command(self, bot, update):
        commands = map(
            lambda key: f'/{key} - {self.registered_commands[key]}',
            self.registered_commands.keys()
        )
        commands = '\n'.join(commands)
        update.message.reply_text(
            'The commands you can execute are: \n\n{commands}'.format(
                name=update.message.from_user.first_name,
                commands=commands
            )
        )

    @bot_command(name='exec', description='Execute a bash command')
    @admin_required
    def bash(self, bot, update):

        message = update.message.text.replace('/exec', '')
        command = os.popen(message).read()
        update.message.reply_text(f'$ {message}\n{command}')


if __name__ == '__main__':
    app = Bot(
        token=os.environ.get('API_TOKEN_KEY', config.API_TOKEN_KEY),
        admins=os.environ.get('ADMINS', config.ADMINS)
    )
    app.run()
