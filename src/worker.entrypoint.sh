#!/bin/bash

# turn on bash's job control
set -m

# Pre-fetch the whisperx VAD model. Upstream whisperx 3.x hardcodes an S3 URL
# that returns 301 and urllib can't follow it (https://github.com/m-bain/whisperX
# bug). Workaround: place the file at the expected torch-hub path so the
# urlopen call is skipped. Model is pyannote/segmentation — user-gated
# (accept terms at https://huggingface.co/pyannote/segmentation).
VAD_FP="$HOME/.cache/torch/whisperx-vad-segmentation.bin"
if [ ! -f "$VAD_FP" ]; then
  echo "Pre-downloading whisperx VAD model from HF..."
  mkdir -p "$(dirname "$VAD_FP")"
  if curl -fL -H "Authorization: Bearer $WHISPER_HF_TOKEN" \
       https://huggingface.co/pyannote/segmentation/resolve/main/pytorch_model.bin \
       -o "$VAD_FP"; then
    echo "VAD model cached at $VAD_FP"
  else
    echo "ERROR: VAD download failed. Accept terms at https://huggingface.co/pyannote/segmentation" >&2
    rm -f "$VAD_FP"
  fi
fi

echo -n "
_______                             ______       ___       __            ______
___    |___________  ____________  ____  /_      __ |     / /_______________  /______________
__  /| |_  __ \_  / / /_  ___/  / / /_  __ \     __ | /| / /_  __ \_  ___/_  //_/  _ \_  ___/
_  ___ |  / / /  /_/ /_(__  )/ /_/ /_  /_/ /     __ |/ |/ / / /_/ /  /   _  ,<  /  __/  /
/_/  |_/_/ /_/_\__, / /____/ \__,_/ /_.___/      ____/|__/  \____//_/    /_/|_| \___//_/
              /____/
"

# Start anysub task server in the background
echo "running task server"
/usr/bin/anysub --task-server &

# Execute the base image's default CMD
echo "running default command"
cd /app
python3 main.py &

# Wait for any process to exit
wait -n

# Exit with status of process that exited first
exit $?
