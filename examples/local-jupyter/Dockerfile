FROM quay.io/jupyter/minimal-notebook:2024-12-23

RUN pip install openai jupyter-ai[all]

CMD [ "start-notebook.sh", "--NotebookApp.token=''", "-ServerApp.root_dir=/home/jovyan/work" ]