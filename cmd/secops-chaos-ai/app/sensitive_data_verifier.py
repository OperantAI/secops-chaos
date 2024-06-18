from presidio_analyzer import AnalyzerEngine
from presidio_analyzer.nlp_engine import TransformersNlpEngine
from .verifiers import AIExperimentVerifierResult

def VerifySensitiveData(check: str, data: str):

    # Define which transformers model to use
    model_config = [{"lang_code": "en", "model_name": {
        "spacy": "en_core_web_sm",  # use a small spaCy model for lemmas, tokens etc.
        "transformers": "dslim/bert-base-NER"
        }
    }]

    nlp_engine = TransformersNlpEngine(models=model_config)

    # Set up the engine, loads the NLP module (spaCy model by default)
    # and other PII recognizers
    analyzer = AnalyzerEngine(nlp_engine=nlp_engine)

    # Call analyzer to get results
    results = analyzer.analyze(text=data, language='en')
    print(data, results)
    verified_results = list()
    for i in results:
        verified_results.append(AIExperimentVerifierResult(check=check, entityType=i.entity_type, detected=bool(True), score = i.score))
    return verified_results