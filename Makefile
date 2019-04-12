.PHONY: build clean start

build:
	@python -m venv .venv;
	. .venv/bin/activate; pip install -r requirements.txt

clean:
	@rm -rf .venv
	@find -iname "*.pyc" -delete

start: 
	@. .venv/bin/activate; python -m src.main
